package runtime

import "sync"

type pendingAudioCallback struct {
	id      uint32
	oneShot bool
}

var audioCallbacks = struct {
	next        uint32
	items       map[uint32]func()
	queue       [32]pendingAudioCallback
	head, count uint8
	mu          sync.Mutex
}{items: make(map[uint32]func())}

// RegisterAudioCallback retains a callback outside native audio userdata.
func RegisterAudioCallback(callback func()) uint32 {
	if callback == nil {
		return 0
	}
	audioCallbacks.mu.Lock()
	for {
		audioCallbacks.next++
		if audioCallbacks.next != 0 && audioCallbacks.items[audioCallbacks.next] == nil {
			audioCallbacks.items[audioCallbacks.next] = callback
			audioCallbacks.mu.Unlock()
			return audioCallbacks.next
		}
	}
}

// InvokeAudioCallback runs a retained callback and optionally releases it.
func InvokeAudioCallback(id uint32, oneShot bool) {
	if id == 0 {
		return
	}
	audioCallbacks.mu.Lock()
	if audioCallbacks.count == uint8(len(audioCallbacks.queue)) {
		if oneShot {
			delete(audioCallbacks.items, id)
		}
		audioCallbacks.mu.Unlock()
		return
	}
	index := (audioCallbacks.head + audioCallbacks.count) % uint8(len(audioCallbacks.queue))
	audioCallbacks.queue[index] = pendingAudioCallback{id: id, oneShot: oneShot}
	audioCallbacks.count++
	audioCallbacks.mu.Unlock()
}

// ForgetAudioCallback releases a callback without invoking it.
func ForgetAudioCallback(id uint32) {
	audioCallbacks.mu.Lock()
	delete(audioCallbacks.items, id)
	audioCallbacks.mu.Unlock()
}

// DrainAudioCallbacks runs queued native completions on the frame-update
// goroutine. Native audio callbacks only enqueue fixed-size records.
func DrainAudioCallbacks() {
	for {
		audioCallbacks.mu.Lock()
		if audioCallbacks.count == 0 {
			audioCallbacks.mu.Unlock()
			return
		}
		entry := audioCallbacks.queue[audioCallbacks.head]
		audioCallbacks.queue[audioCallbacks.head] = pendingAudioCallback{}
		audioCallbacks.head = (audioCallbacks.head + 1) % uint8(len(audioCallbacks.queue))
		audioCallbacks.count--
		callback := audioCallbacks.items[entry.id]
		if entry.oneShot {
			delete(audioCallbacks.items, entry.id)
		}
		audioCallbacks.mu.Unlock()
		if callback != nil {
			callback()
		}
	}
}
