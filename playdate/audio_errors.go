package playdate

// Audio errors.

type audioError string

func (message audioError) Error() string { return string(message) }

// AudioLoadError identifies an audio asset that Playdate could not load.
type AudioLoadError string

func (path AudioLoadError) Error() string { return "load audio: " + string(path) }

var (
	// ErrAudioClosed indicates an operation on a closed audio player.
	ErrAudioClosed error = audioError("audio player is closed")
	// ErrAudioCreate indicates that Playdate could not allocate a player.
	ErrAudioCreate error = audioError("create audio player failed")
	// ErrAudioPlay indicates that Playdate rejected playback.
	ErrAudioPlay error = audioError("audio playback failed")
	// ErrAudioVolume indicates a non-finite volume outside the 0..1 range.
	ErrAudioVolume error = audioError("audio volume must be between zero and one")
	// ErrAudioRepeat indicates a repeat count outside the native 32-bit range.
	ErrAudioRepeat error = audioError("audio repeat count must be between zero and 2147483647")
	// ErrAudioRate indicates a zero or non-finite playback rate.
	ErrAudioRate error = audioError("audio rate must be finite and non-zero")
	// ErrAudioReverseUnsupported indicates reverse playback on a streaming player.
	ErrAudioReverseUnsupported error = audioError("streaming audio cannot play in reverse")
	// ErrAudioOffset indicates a non-finite or negative playback offset.
	ErrAudioOffset error = audioError("audio offset must be finite and non-negative")
	// ErrAudioFade indicates a fade duration outside the native signed 32-bit range.
	ErrAudioFade error = audioError("audio fade duration must not exceed 2147483647 frames")
	// ErrAudioUnavailable indicates that an optional advanced audio capability is absent.
	ErrAudioUnavailable error = audioError("advanced audio is unavailable")
	// ErrAudioChannelClosed indicates an operation on a closed routing channel.
	ErrAudioChannelClosed error = audioError("audio channel is closed")
	// ErrAudioSourceInvalid indicates a source not created by this runtime.
	ErrAudioSourceInvalid error = audioError("audio source is invalid")
	// ErrAudioRoute indicates that Playdate rejected a routing graph change.
	ErrAudioRoute error = audioError("audio routing change failed")
	// ErrAudioPan indicates a non-finite pan outside the -1..1 range.
	ErrAudioPan error = audioError("audio pan must be between negative one and one")
	// ErrAudioGraphClosed indicates an operation on a closed synth or signal.
	ErrAudioGraphClosed error = audioError("audio graph node is closed")
	// ErrAudioParameter indicates an invalid non-finite synth or signal value.
	ErrAudioParameter error = audioError("audio parameter must be finite")
	// ErrAudioWaveform indicates a waveform outside the native enumeration.
	ErrAudioWaveform error = audioError("audio waveform is invalid")
	// ErrAudioEventStep indicates a negative control-signal step.
	ErrAudioEventStep error = audioError("audio event step must be non-negative")
)
