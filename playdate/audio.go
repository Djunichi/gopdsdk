package playdate

// PlaybackState describes the observable state of an audio player.
type PlaybackState uint8

const (
	PlaybackStopped PlaybackState = iota
	PlaybackPlaying
	PlaybackPaused
)

// SoundEffect is an explicitly owned, memory-backed short sound.
type SoundEffect interface {
	Play() error
	Stop() error
	SetVolume(left, right float32) error
	Volume() (left, right float32, err error)
	State() (PlaybackState, error)
	Pause() error
	Resume() error
	Close() error
}

// SamplePlayer is an explicitly owned, memory-backed player with playback
// controls beyond the SoundEffect convenience API. Times are in seconds.
type SamplePlayer interface {
	SoundEffect
	PlayRepeated(repeat int, rate float32) error
	Length() (float32, error)
	SetOffset(seconds float32) error
	Offset() (float32, error)
	SetRate(rate float32) error
	Rate() (float32, error)
}

// SamplePlayers loads advanced memory-backed sample players. Games should
// capability-assert this optional interface from Context.
type SamplePlayers interface {
	LoadSamplePlayer(path string) (SamplePlayer, error)
}

// VariableRatePlayer changes and reports a player's playback rate. FilePlayer
// values may capability-assert this optional interface, but only positive rates
// are supported for streaming playback.
type VariableRatePlayer interface {
	SetRate(rate float32) error
	Rate() (float32, error)
}

// FilePlayer is an explicitly owned streaming audio player for one file.
type FilePlayer interface {
	Play() error
	Stop() error
	SetVolume(left, right float32) error
	Volume() (left, right float32, err error)
	State() (PlaybackState, error)
	Pause() error
	Resume() error
	Close() error
}

// Audio loads the two accepted audio use cases.
type Audio interface {
	LoadSoundEffect(path string) (SoundEffect, error)
	LoadFilePlayer(path string) (FilePlayer, error)
}
