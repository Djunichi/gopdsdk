package schedule

// Queue is a bounded non-blocking FIFO. Its storage is allocated by NewQueue;
// sending and receiving do not allocate. Queue is not safe for concurrent use.
type Queue[T any] struct {
	values []T
	head   int
	length int
}

// NewQueue creates a fixed-capacity queue.
func NewQueue[T any](capacity int) (*Queue[T], error) {
	if capacity <= 0 {
		return nil, ErrConfig
	}
	return &Queue[T]{values: make([]T, capacity)}, nil
}

// TrySend appends value or returns false without waiting when the queue is full.
func (queue *Queue[T]) TrySend(value T) bool {
	if queue.length == len(queue.values) {
		return false
	}
	queue.values[(queue.head+queue.length)%len(queue.values)] = value
	queue.length++
	return true
}

// TryReceive removes the oldest value or returns false without waiting when empty.
func (queue *Queue[T]) TryReceive() (T, bool) {
	if queue.length == 0 {
		var zero T
		return zero, false
	}
	value := queue.values[queue.head]
	var zero T
	queue.values[queue.head] = zero
	queue.head = (queue.head + 1) % len(queue.values)
	queue.length--
	return value, true
}

// Len returns the current occupancy.
func (queue *Queue[T]) Len() int { return queue.length }

// Capacity returns the fixed maximum occupancy.
func (queue *Queue[T]) Capacity() int { return len(queue.values) }

// Clear removes all values and releases references retained in queue storage.
func (queue *Queue[T]) Clear() {
	var zero T
	for queue.length > 0 {
		queue.values[queue.head] = zero
		queue.head = (queue.head + 1) % len(queue.values)
		queue.length--
	}
	queue.head = 0
}

// Poller checks fixed queues with deterministic round-robin priority.
type Poller[T any] struct {
	queues []*Queue[T]
	next   int
}

// NewPoller copies and validates a non-empty queue set.
func NewPoller[T any](queues ...*Queue[T]) (*Poller[T], error) {
	if len(queues) == 0 {
		return nil, ErrConfig
	}
	owned := append([]*Queue[T](nil), queues...)
	for _, queue := range owned {
		if queue == nil {
			return nil, ErrConfig
		}
	}
	return &Poller[T]{queues: owned}, nil
}

// TryReceive polls each queue once, beginning at the priority after the last
// successful queue. It returns the selected queue index and an explicit ready result.
func (poller *Poller[T]) TryReceive() (value T, queueIndex int, ready bool) {
	for offset := 0; offset < len(poller.queues); offset++ {
		index := (poller.next + offset) % len(poller.queues)
		if value, ready = poller.queues[index].TryReceive(); ready {
			poller.next = (index + 1) % len(poller.queues)
			return value, index, true
		}
	}
	return value, -1, false
}
