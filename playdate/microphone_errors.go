package playdate

// Microphone errors.

type microphoneError string

func (message microphoneError) Error() string { return string(message) }

var (
	// ErrMicrophoneUnavailable indicates a context without microphone input.
	ErrMicrophoneUnavailable error = microphoneError("microphone capability is unavailable")
	// ErrMicrophoneCallback indicates a missing permission or recording callback.
	ErrMicrophoneCallback error = microphoneError("microphone callback is required")
	// ErrMicrophoneSource indicates an invalid input-source selection.
	ErrMicrophoneSource error = microphoneError("microphone source is invalid")
	// ErrMicrophoneDenied indicates that microphone permission was denied.
	ErrMicrophoneDenied error = microphoneError("microphone access was denied")
	// ErrMicrophoneStart indicates that native recording could not start.
	ErrMicrophoneStart error = microphoneError("microphone recording failed to start")
	// ErrMicrophoneClosed indicates an operation on a closed recorder.
	ErrMicrophoneClosed error = microphoneError("microphone recorder is closed")
	// ErrMicrophoneSamplesExpired indicates access outside the native callback.
	ErrMicrophoneSamplesExpired error = microphoneError("microphone samples have expired")
)
