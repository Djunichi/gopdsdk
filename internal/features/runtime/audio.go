package runtime

var audioCallbacks = struct {
	next  uint32
	items map[uint32]func()
}{items: make(map[uint32]func())}

// RegisterAudioCallback retains a callback outside native audio userdata.
func RegisterAudioCallback(callback func()) uint32 {
	if callback == nil {
		return 0
	}
	for {
		audioCallbacks.next++
		if audioCallbacks.next != 0 && audioCallbacks.items[audioCallbacks.next] == nil {
			audioCallbacks.items[audioCallbacks.next] = callback
			return audioCallbacks.next
		}
	}
}

// InvokeAudioCallback runs a retained callback and optionally releases it.
func InvokeAudioCallback(id uint32, oneShot bool) {
	callback := audioCallbacks.items[id]
	if oneShot {
		delete(audioCallbacks.items, id)
	}
	if callback != nil {
		callback()
	}
}

// ForgetAudioCallback releases a callback without invoking it.
func ForgetAudioCallback(id uint32) { delete(audioCallbacks.items, id) }
