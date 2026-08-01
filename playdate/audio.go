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
