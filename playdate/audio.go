package playdate

// PlaybackState describes the observable state of an audio player.
type PlaybackState uint8

const (
	PlaybackStopped PlaybackState = iota
	PlaybackPlaying
	PlaybackPaused
)

// AudioSource is a player that can be routed through an AudioChannel. The
// player retains its own lifetime when attached to a channel.
type AudioSource interface {
	SetVolume(left, right float32) error
	Volume() (left, right float32, err error)
	State() (PlaybackState, error)
}

// SoundEffect is an explicitly owned, memory-backed short sound.
type SoundEffect interface {
	AudioSource
	Play() error
	Stop() error
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

// CompletionPlayer reports natural playback completion. Setting a nil callback
// clears the current callback. Stop and Close do not invoke it.
type CompletionPlayer interface {
	SetFinishCallback(callback func()) error
}

// FadingPlayer performs a linear volume fade over audioFrames. The optional
// callback runs when the fade completes. FilePlayer values may capability-
// assert this interface.
type FadingPlayer interface {
	FadeVolume(left, right float32, audioFrames uint32, callback func()) error
}

// AudioClock exposes Playdate's wrapping 44.1 kHz audio-frame clock. Games
// should capability-assert this optional interface from Context.
type AudioClock interface {
	CurrentAudioTime() (uint32, error)
}

// AudioChannel owns one native routing node and its source attachments. Closing
// a source detaches it; closing the channel detaches every source. Neither
// operation closes the other owned object.
type AudioChannel interface {
	AddSource(source AudioSource) error
	RemoveSource(source AudioSource) error
	SetVolume(volume float32) error
	Volume() (float32, error)
	SetPan(pan float32) error
	Close() error
}

// AudioChannels creates explicitly owned audio routing channels. Games should
// capability-assert this optional interface from Context.
type AudioChannels interface {
	NewAudioChannel() (AudioChannel, error)
}

// Waveform selects a native oscillator shape.
type Waveform uint8

const (
	WaveformSquare Waveform = iota
	WaveformTriangle
	WaveformSine
	WaveformNoise
	WaveformSawtooth
	WaveformPOPhase
	WaveformPODigital
	WaveformPOVosim
)

// LFOType selects a native low-frequency oscillator shape.
type LFOType uint8

const (
	LFOTypeSquare LFOType = iota
	LFOTypeTriangle
	LFOTypeSine
	LFOTypeSampleAndHold
	LFOTypeSawtoothUp
	LFOTypeSawtoothDown
	LFOTypeArpeggiator
)

// Signal is an explicitly owned modulation value. Attachments do not transfer
// ownership and are detached when either endpoint closes.
type Signal interface {
	Value() (float32, error)
	SetScale(scale float32) error
	SetOffset(offset float32) error
	Close() error
}

// LFO is a periodic modulation signal. Rate is measured in cycles per second;
// phase, center, and depth use native Playdate units.
type LFO interface {
	Signal
	SetRate(rate float32) error
	SetPhase(phase float32) error
	SetCenter(center float32) error
	SetDepth(depth float32) error
	SetRetrigger(retrigger bool) error
}

// Envelope is an ADSR modulation signal whose times are measured in seconds.
type Envelope interface {
	Signal
	SetAttack(seconds float32) error
	SetDecay(seconds float32) error
	SetSustain(level float32) error
	SetRelease(seconds float32) error
	SetLegato(legato bool) error
	SetRetrigger(retrigger bool) error
}

// ControlSignal is a step timeline usable as a synth modulator.
type ControlSignal interface {
	Signal
	AddEvent(step int, value float32, interpolate bool) error
	RemoveEvent(step int) error
	ClearEvents() error
}

// Synth is an explicitly owned native oscillator and routable audio source.
// Note scheduling uses the wrapping 44.1 kHz audio-frame clock.
type Synth interface {
	AudioSource
	SetWaveform(waveform Waveform) error
	SetEnvelope(attack, decay, sustain, release float32) error
	SetTranspose(semitones float32) error
	SetFrequencyModulator(signal Signal) error
	SetAmplitudeModulator(signal Signal) error
	PlayMIDINote(note, velocity, length float32, when uint32) error
	NoteOff(when uint32) error
	Stop() error
	Close() error
}

// Synthesizers creates owned synth and modulation graph nodes. Games should
// capability-assert this optional interface from Context.
type Synthesizers interface {
	NewSynth(waveform Waveform) (Synth, error)
	NewLFO(lfoType LFOType) (LFO, error)
	NewEnvelope(attack, decay, sustain, release float32) (Envelope, error)
	NewControlSignal() (ControlSignal, error)
}

// FilePlayer is an explicitly owned streaming audio player for one file.
type FilePlayer interface {
	AudioSource
	Play() error
	Stop() error
	Pause() error
	Resume() error
	Close() error
}

// Audio loads the two accepted audio use cases.
type Audio interface {
	LoadSoundEffect(path string) (SoundEffect, error)
	LoadFilePlayer(path string) (FilePlayer, error)
}
