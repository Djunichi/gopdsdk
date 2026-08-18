package schedule

import (
	"errors"
	"reflect"
	"testing"
)

type testClock struct{ now uint32 }

func (clock *testClock) CurrentTimeMilliseconds() uint32 { return clock.now }

func newTestScheduler(t *testing.T, clock *testClock, capacity, steps int) *Scheduler {
	t.Helper()
	scheduler, err := New(clock, Config{Capacity: capacity, StepsPerFrame: steps})
	if err != nil {
		t.Fatal(err)
	}
	return scheduler
}

func TestSchedulerFIFOAndOneStepPerFrame(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 3, 3)
	var order []int
	for task := 1; task <= 3; task++ {
		task := task
		calls := 0
		if _, err := scheduler.Schedule(func() Action {
			order = append(order, task)
			calls++
			if calls == 2 {
				return Complete()
			}
			return Yield()
		}); err != nil {
			t.Fatal(err)
		}
	}
	first, err := scheduler.Update()
	if err != nil || first.Steps != 3 || first.Pending != 3 {
		t.Fatalf("first frame = %+v, %v", first, err)
	}
	second, err := scheduler.Update()
	if err != nil || second.Completed != 3 || second.Pending != 0 {
		t.Fatalf("second frame = %+v, %v", second, err)
	}
	if want := []int{1, 2, 3, 1, 2, 3}; !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
}

func TestSchedulerStepBoundAndSchedulingDuringStep(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 3, 1)
	var order []int
	_, err := scheduler.Schedule(func() Action {
		order = append(order, 1)
		if _, scheduleErr := scheduler.Schedule(func() Action {
			order = append(order, 2)
			return Complete()
		}); scheduleErr != nil {
			t.Fatal(scheduleErr)
		}
		return Complete()
	})
	if err != nil {
		t.Fatal(err)
	}
	if frame, updateErr := scheduler.Update(); updateErr != nil || frame.Steps != 1 || !reflect.DeepEqual(order, []int{1}) {
		t.Fatalf("first frame = %+v, %v, order %v", frame, updateErr, order)
	}
	if frame, updateErr := scheduler.Update(); updateErr != nil || frame.Steps != 1 || !reflect.DeepEqual(order, []int{1, 2}) {
		t.Fatalf("second frame = %+v, %v, order %v", frame, updateErr, order)
	}
}

func TestSchedulerCancellationDuringStepDoesNotCorruptFrameQueue(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 4, 4)
	var order []int
	var cancelled Task
	if _, err := scheduler.Schedule(func() Action {
		order = append(order, 1)
		if !scheduler.Cancel(cancelled) {
			t.Fatal("cancel queued task failed")
		}
		if _, scheduleErr := scheduler.Schedule(func() Action {
			order = append(order, 4)
			return Complete()
		}); scheduleErr != nil {
			t.Fatal(scheduleErr)
		}
		return Complete()
	}); err != nil {
		t.Fatal(err)
	}
	cancelled, _ = scheduler.Schedule(func() Action {
		order = append(order, 2)
		return Complete()
	})
	_, _ = scheduler.Schedule(func() Action {
		order = append(order, 3)
		return Complete()
	})
	frame, err := scheduler.Update()
	if err != nil || !reflect.DeepEqual(order, []int{1, 3}) || frame.Pending != 1 {
		t.Fatalf("first frame = %+v, %v, order %v", frame, err, order)
	}
	frame, err = scheduler.Update()
	if err != nil || !reflect.DeepEqual(order, []int{1, 3, 4}) || frame.Pending != 0 {
		t.Fatalf("second frame = %+v, %v, order %v", frame, err, order)
	}
}

func TestSchedulerSelfCancelAndNestedUpdate(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 1, 1)
	var task Task
	var nested error
	task, _ = scheduler.Schedule(func() Action {
		_, nested = scheduler.Update()
		if !scheduler.Cancel(task) {
			t.Fatal("self cancel failed")
		}
		return Yield()
	})
	frame, err := scheduler.Update()
	if err != nil || !errors.Is(nested, ErrUpdating) || frame.Pending != 0 {
		t.Fatalf("frame = %+v, err %v, nested %v", frame, err, nested)
	}
}

func TestSchedulerDelayDeadlineRepeatAndClockWrap(t *testing.T) {
	clock := &testClock{now: ^uint32(0) - 4}
	scheduler := newTestScheduler(t, clock, 3, 3)
	var calls []string
	if _, err := scheduler.ScheduleAfter(10, func() Action {
		calls = append(calls, "after")
		return Complete()
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := scheduler.ScheduleAt(clock.now+5, func() Action {
		calls = append(calls, "at")
		return Complete()
	}); err != nil {
		t.Fatal(err)
	}
	repeats := 0
	if _, err := scheduler.Schedule(func() Action {
		repeats++
		calls = append(calls, "repeat")
		if repeats == 3 {
			return Complete()
		}
		return Repeat(4)
	}); err != nil {
		t.Fatal(err)
	}
	_, _ = scheduler.Update()
	clock.now += 4
	_, _ = scheduler.Update()
	clock.now += 1
	_, _ = scheduler.Update()
	clock.now += 3
	_, _ = scheduler.Update()
	clock.now += 2
	frame, err := scheduler.Update()
	if err != nil || frame.Pending != 0 {
		t.Fatalf("last frame = %+v, %v", frame, err)
	}
	if want := []string{"repeat", "repeat", "at", "repeat", "after"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestSchedulerCancellationCapacityAndStaleHandles(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 1, 1)
	first, err := scheduler.Schedule(func() Action { return Yield() })
	if err != nil {
		t.Fatal(err)
	}
	if _, err = scheduler.Schedule(func() Action { return Complete() }); !errors.Is(err, ErrFull) {
		t.Fatalf("full error = %v", err)
	}
	if !scheduler.Cancel(first) || scheduler.Pending(first) || scheduler.Cancel(first) {
		t.Fatal("unexpected cancellation lifecycle")
	}
	second, err := scheduler.Schedule(func() Action { return Complete() })
	if err != nil || second == first || scheduler.Pending(first) {
		t.Fatalf("reused task = %v, %v", second, err)
	}
	if _, err = scheduler.Update(); err != nil {
		t.Fatal(err)
	}
}

func TestSchedulerBudgetAndInvalidAction(t *testing.T) {
	clock := &testClock{}
	scheduler, err := New(clock, Config{Capacity: 2, StepsPerFrame: 2, TimeBudgetMilliseconds: 2})
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 2; index++ {
		if _, err = scheduler.Schedule(func() Action {
			clock.now += 2
			return Yield()
		}); err != nil {
			t.Fatal(err)
		}
	}
	frame, err := scheduler.Update()
	if err != nil || frame.Steps != 1 || !frame.BudgetExhausted {
		t.Fatalf("budget frame = %+v, %v", frame, err)
	}

	bad := newTestScheduler(t, clock, 1, 1)
	if _, err = bad.Schedule(func() Action { return Repeat(0) }); err != nil {
		t.Fatal(err)
	}
	frame, err = bad.Update()
	if !errors.Is(err, ErrDelay) || frame.Pending != 0 {
		t.Fatalf("invalid repeat = %+v, %v", frame, err)
	}
}

func TestSchedulerTerminateReleasesTasks(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 1, 1)
	task, err := scheduler.Schedule(func() Action { return Yield() })
	if err != nil {
		t.Fatal(err)
	}
	scheduler.Terminate()
	scheduler.Terminate()
	if scheduler.Pending(task) {
		t.Fatal("task remains pending")
	}
	if _, err = scheduler.Update(); !errors.Is(err, ErrTerminated) {
		t.Fatalf("update error = %v", err)
	}
	if _, err = scheduler.Schedule(func() Action { return Complete() }); !errors.Is(err, ErrTerminated) {
		t.Fatalf("schedule error = %v", err)
	}
}

func TestSchedulerUpdateDoesNotAllocate(t *testing.T) {
	clock := &testClock{}
	scheduler := newTestScheduler(t, clock, 1, 1)
	if _, err := scheduler.Schedule(func() Action { return Yield() }); err != nil {
		t.Fatal(err)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		if _, err := scheduler.Update(); err != nil {
			t.Fatal(err)
		}
	}); allocations != 0 {
		t.Fatalf("allocations per update = %v", allocations)
	}
}
