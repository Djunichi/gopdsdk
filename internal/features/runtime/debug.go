package runtime

// DebugMessageQueue is a bounded FIFO used by native serial callbacks. It is
// intentionally single-threaded because Playdate callbacks run on the game
// event thread.
type DebugMessageQueue struct {
	items      []string
	capacity   int
	maxMessage int
}

// NewDebugMessageQueue creates a queue with fixed message-count and byte
// bounds. Non-positive bounds create a queue that discards all messages.
func NewDebugMessageQueue(capacity, maxMessage int) *DebugMessageQueue {
	return &DebugMessageQueue{capacity: capacity, maxMessage: maxMessage}
}

// Push copies a message into the queue, truncating it to the configured byte
// bound and dropping the oldest item when full.
func (queue *DebugMessageQueue) Push(message string) {
	if queue.capacity <= 0 || queue.maxMessage <= 0 {
		return
	}
	if len(message) > queue.maxMessage {
		message = message[:queue.maxMessage]
	}
	if len(queue.items) == queue.capacity {
		copy(queue.items, queue.items[1:])
		queue.items = queue.items[:queue.capacity-1]
	}
	queue.items = append(queue.items, message)
}

// Poll removes the oldest queued message.
func (queue *DebugMessageQueue) Poll() (string, bool) {
	if len(queue.items) == 0 {
		return "", false
	}
	message := queue.items[0]
	copy(queue.items, queue.items[1:])
	queue.items = queue.items[:len(queue.items)-1]
	return message, true
}

// Clear releases all queued messages.
func (queue *DebugMessageQueue) Clear() { queue.items = nil }
