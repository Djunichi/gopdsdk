// Package schedule provides bounded cooperative tasks driven by Playdate's
// wrapping millisecond clock.
package schedule

// Clock supplies the wrapping monotonic device clock used by Scheduler.
type Clock interface {
	CurrentTimeMilliseconds() uint32
}

// Step is one small, explicitly yielding unit of task work.
type Step func() Action

// Task identifies one scheduled task. The zero value never identifies a task.
type Task uint64

// Config fixes the scheduler's storage and per-frame work bounds.
type Config struct {
	Capacity               int
	StepsPerFrame          int
	TimeBudgetMilliseconds uint32
}

// Frame reports the work performed by one Update call.
type Frame struct {
	Steps           int
	Completed       int
	Pending         int
	BudgetExhausted bool
}

type actionKind uint8

const (
	actionYield actionKind = iota
	actionComplete
	actionDelay
	actionRepeat
)

// Action tells the scheduler what to do after a step returns.
type Action struct {
	kind         actionKind
	milliseconds uint32
}

// Yield runs the task again on a later update frame.
func Yield() Action { return Action{kind: actionYield} }

// Complete finishes the task and releases its scheduler slot.
func Complete() Action { return Action{kind: actionComplete} }

// Delay runs the task again after at least milliseconds on the wrapping
// Playdate clock. Durations greater than MaxDelayMilliseconds are invalid.
func Delay(milliseconds uint32) Action {
	return Action{kind: actionDelay, milliseconds: milliseconds}
}

// Repeat runs the task after at least milliseconds and keeps the cadence
// anchored to its previous due time. It is useful for bounded ticker work.
func Repeat(milliseconds uint32) Action {
	return Action{kind: actionRepeat, milliseconds: milliseconds}
}

// MaxDelayMilliseconds is the largest unambiguous interval on a wrapping
// uint32 millisecond clock.
const MaxDelayMilliseconds = uint32(1<<31 - 1)

type taskState uint8

const (
	taskFree taskState = iota
	taskQueued
)

type taskSlot struct {
	step       Step
	due        uint32
	generation uint32
	state      taskState
	repeating  bool
}

// Scheduler owns fixed-capacity task and ready-order storage. Call Update only
// from the application's update callback. A task runs at most once per Update.
type Scheduler struct {
	clock      Clock
	config     Config
	tasks      []taskSlot
	queue      []int
	head       int
	length     int
	pending    int
	terminated bool
	updating   bool
	remaining  int
}

// New validates config and allocates all scheduler-owned frame storage.
func New(clock Clock, config Config) (*Scheduler, error) {
	if clock == nil || config.Capacity <= 0 || config.StepsPerFrame <= 0 || config.TimeBudgetMilliseconds > MaxDelayMilliseconds {
		return nil, ErrConfig
	}
	return &Scheduler{
		clock:  clock,
		config: config,
		tasks:  make([]taskSlot, config.Capacity),
		queue:  make([]int, config.Capacity),
	}, nil
}

// Schedule appends a task to the deterministic FIFO update order.
func (scheduler *Scheduler) Schedule(step Step) (Task, error) {
	return scheduler.scheduleAt(step, scheduler.clock.CurrentTimeMilliseconds())
}

// ScheduleAfter appends a task that becomes ready after the given interval.
func (scheduler *Scheduler) ScheduleAfter(milliseconds uint32, step Step) (Task, error) {
	if milliseconds > MaxDelayMilliseconds {
		return 0, ErrDelay
	}
	return scheduler.scheduleAt(step, scheduler.clock.CurrentTimeMilliseconds()+milliseconds)
}

// ScheduleAt appends a task for an absolute wrapping-clock deadline. The
// deadline must be no more than MaxDelayMilliseconds into the future.
func (scheduler *Scheduler) ScheduleAt(deadline uint32, step Step) (Task, error) {
	now := scheduler.clock.CurrentTimeMilliseconds()
	if !reached(now, deadline) && deadline-now > MaxDelayMilliseconds {
		return 0, ErrDelay
	}
	return scheduler.scheduleAt(step, deadline)
}

func (scheduler *Scheduler) scheduleAt(step Step, deadline uint32) (Task, error) {
	if scheduler.terminated {
		return 0, ErrTerminated
	}
	if step == nil {
		return 0, ErrTask
	}
	if scheduler.pending == len(scheduler.tasks) {
		return 0, ErrFull
	}
	for index := range scheduler.tasks {
		slot := &scheduler.tasks[index]
		if slot.state != taskFree {
			continue
		}
		slot.generation++
		if slot.generation == 0 {
			slot.generation++
		}
		slot.step = step
		slot.due = deadline
		slot.state = taskQueued
		slot.repeating = false
		scheduler.push(index)
		scheduler.pending++
		return makeTask(index, slot.generation), nil
	}
	return 0, ErrFull
}

// Cancel releases a live task. It returns false for completed, cancelled,
// unknown, or stale task identifiers.
func (scheduler *Scheduler) Cancel(task Task) bool {
	index, _, ok := scheduler.identify(task)
	if !ok {
		return false
	}
	if scheduler.removeQueued(index) && scheduler.updating {
		scheduler.remaining--
	}
	scheduler.release(index)
	return true
}

func (scheduler *Scheduler) removeQueued(target int) bool {
	kept := 0
	count := scheduler.length
	removed := false
	for offset := 0; offset < count; offset++ {
		index := scheduler.queue[(scheduler.head+offset)%len(scheduler.queue)]
		if index == target {
			removed = true
			continue
		}
		scheduler.queue[(scheduler.head+kept)%len(scheduler.queue)] = index
		kept++
	}
	scheduler.length = kept
	return removed
}

// Pending reports whether task is still scheduled.
func (scheduler *Scheduler) Pending(task Task) bool {
	_, _, ok := scheduler.identify(task)
	return ok
}

func (scheduler *Scheduler) identify(task Task) (int, uint32, bool) {
	if task == 0 {
		return 0, 0, false
	}
	index := int(uint32(task)) - 1
	generation := uint32(uint64(task) >> 32)
	if index < 0 || index >= len(scheduler.tasks) {
		return 0, 0, false
	}
	slot := &scheduler.tasks[index]
	return index, generation, slot.state == taskQueued && slot.generation == generation
}

// Update runs at most the configured number of ready task steps. TimeBudgetMilliseconds,
// when non-zero, is a secondary guard checked between steps; one step can still
// overrun a frame and must remain small.
func (scheduler *Scheduler) Update() (Frame, error) {
	if scheduler.terminated {
		return Frame{}, ErrTerminated
	}
	if scheduler.updating {
		return Frame{Pending: scheduler.pending}, ErrUpdating
	}
	scheduler.updating = true
	start := scheduler.clock.CurrentTimeMilliseconds()
	now := start
	scheduler.remaining = scheduler.length
	frame := Frame{}
	for scheduler.remaining > 0 && frame.Steps < scheduler.config.StepsPerFrame {
		if frame.Steps > 0 && scheduler.config.TimeBudgetMilliseconds != 0 {
			now = scheduler.clock.CurrentTimeMilliseconds()
			if uint32(now-start) >= scheduler.config.TimeBudgetMilliseconds {
				frame.BudgetExhausted = true
				break
			}
		}
		index := scheduler.pop()
		scheduler.remaining--
		slot := &scheduler.tasks[index]
		if slot.state != taskQueued {
			continue
		}
		if frame.Steps == 0 || scheduler.config.TimeBudgetMilliseconds == 0 {
			now = scheduler.clock.CurrentTimeMilliseconds()
		}
		if !reached(now, slot.due) {
			scheduler.push(index)
			continue
		}
		previousDue := slot.due
		generation := slot.generation
		action := slot.step()
		frame.Steps++
		if scheduler.terminated {
			scheduler.updating = false
			return Frame{Steps: frame.Steps}, ErrTerminated
		}
		if slot.state != taskQueued || slot.generation != generation {
			continue
		}
		switch action.kind {
		case actionComplete:
			scheduler.release(index)
			frame.Completed++
		case actionDelay, actionRepeat:
			if action.milliseconds > MaxDelayMilliseconds || (action.kind == actionRepeat && action.milliseconds == 0) {
				scheduler.release(index)
				frame.Completed++
				frame.Pending = scheduler.pending
				scheduler.updating = false
				return frame, ErrDelay
			}
			if action.kind == actionRepeat {
				slot.due = previousDue + action.milliseconds
				slot.repeating = true
			} else {
				slot.due = scheduler.clock.CurrentTimeMilliseconds() + action.milliseconds
				slot.repeating = false
			}
			scheduler.push(index)
		default:
			slot.due = now
			slot.repeating = false
			scheduler.push(index)
		}
	}
	frame.Pending = scheduler.pending
	scheduler.updating = false
	return frame, nil
}

// Terminate cancels every task, releases retained step closures, and rejects
// subsequent scheduling and updates. It is safe to call more than once.
func (scheduler *Scheduler) Terminate() {
	for index := range scheduler.tasks {
		scheduler.tasks[index].step = nil
		scheduler.tasks[index].state = taskFree
	}
	scheduler.head = 0
	scheduler.length = 0
	scheduler.pending = 0
	scheduler.remaining = 0
	scheduler.terminated = true
}

func (scheduler *Scheduler) release(index int) {
	slot := &scheduler.tasks[index]
	if slot.state == taskQueued {
		slot.step = nil
		slot.state = taskFree
		slot.repeating = false
		scheduler.pending--
	}
}

func (scheduler *Scheduler) push(index int) {
	position := (scheduler.head + scheduler.length) % len(scheduler.queue)
	scheduler.queue[position] = index
	scheduler.length++
}

func (scheduler *Scheduler) pop() int {
	index := scheduler.queue[scheduler.head]
	scheduler.head = (scheduler.head + 1) % len(scheduler.queue)
	scheduler.length--
	return index
}

func makeTask(index int, generation uint32) Task {
	return Task(uint64(generation)<<32 | uint64(uint32(index+1)))
}

func reached(now, deadline uint32) bool {
	return int32(now-deadline) >= 0
}
