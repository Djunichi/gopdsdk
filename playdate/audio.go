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

// SamplePlayerControls changes the sample and frame range of an owned player.
type SamplePlayerControls interface {
	SetSample(AudioSample) error
	SetPlayRange(startFrame, endFrame int) error
}

// LoopCallbackPlayer reports each completed playback loop. A nil callback
// clears the registration.
type LoopCallbackPlayer interface{ SetLoopCallback(callback func()) error }

// SamplePlayers loads advanced memory-backed sample players. Games should
// capability-assert this optional interface from Context.
type SamplePlayers interface {
	LoadSamplePlayer(path string) (SamplePlayer, error)
}

// SamplePlayerFactory creates an empty player and attaches an existing sample.
type SamplePlayerFactory interface {
	NewSamplePlayer(sample AudioSample) (SamplePlayer, error)
}

// SoundFormat describes the storage layout of an AudioSample.
type SoundFormat uint8

const (
	Sound8BitMono    SoundFormat = 0
	Sound8BitStereo  SoundFormat = 1
	Sound16BitMono   SoundFormat = 2
	Sound16BitStereo SoundFormat = 3
	SoundADPCMMono   SoundFormat = 4
	SoundADPCMStereo SoundFormat = 5
)

// SampleData is a borrowed view of native sample storage. CopyTo is valid only
// while the originating AudioSample remains open and copies at most len(dst)
// bytes without retaining dst.
type SampleData interface {
	Len() int
	Format() SoundFormat
	SampleRate() uint32
	CopyTo(dst []byte) (int, error)
}

// AudioSample owns a native sample buffer. Players borrow attached samples;
// callers must keep a sample open until it is replaced or the player is closed.
type AudioSample interface {
	Load(path string) error
	Data() (SampleData, error)
	Length() (float32, error)
	Decompress() error
	Close() error
}

// AudioSamples creates native-owned samples. NewSampleFromData copies data, so
// the caller may immediately reuse or release its input slice.
type AudioSamples interface {
	NewSample(byteCount int) (AudioSample, error)
	LoadSample(path string) (AudioSample, error)
	NewSampleFromData(data []byte, format SoundFormat, sampleRate uint32) (AudioSample, error)
}

// PCMPlayers copies caller-owned mono signed 16-bit PCM into a native-owned
// sample player. The input slice is not retained after the call returns.
type PCMPlayers interface {
	NewPCMPlayer(samples []int16, sampleRate uint32) (SamplePlayer, error)
}

// PCMRenderCallback fills up to len(left) signed 16-bit PCM frames and returns
// the number produced. right is nil for mono sources. The callback runs only
// from the frame-update goroutine, never from Playdate's audio thread.
type PCMRenderCallback func(left, right []int16) int

// PCMCallbackSource is a continuously playing, explicitly owned source backed
// by a bounded native PCM ring. An underrun emits silence without calling Go.
type PCMCallbackSource interface {
	AudioSource
	UnderrunCount() (uint32, error)
	Close() error
}

// CallbackAudio creates bounded callback sources attached to an existing
// channel. stereo selects whether the render callback receives a right buffer.
type CallbackAudio interface {
	NewPCMCallbackSource(channel AudioChannel, stereo bool, callback PCMRenderCallback) (PCMCallbackSource, error)
}

// GeneratorState is the latest bounded snapshot for one native synth voice.
// Parameters contains the eight custom generator parameter slots.
type GeneratorState struct {
	Voice         uint8
	Note          float32
	Velocity      float32
	Length        float32
	Released      bool
	ReleaseOffset int32
	Rate          uint32
	DeltaRate     int32
	Parameters    [8]float32
}

// GeneratorRenderCallback produces signed 16-bit PCM for one synth voice. The
// callback runs on the frame-update goroutine; right is nil for mono generators.
type GeneratorRenderCallback func(state GeneratorState, left, right []int16) int

// GeneratorSynth is a native PDSynth configured with a device-safe custom
// generator. Instrument voice copies receive independent bounded PCM rings.
type GeneratorSynth interface{ Synth }

// GeneratorSynthesizers creates custom generator synths.
type GeneratorSynthesizers interface {
	NewGeneratorSynth(stereo bool, callback GeneratorRenderCallback) (GeneratorSynth, error)
}

// VariableRatePlayer changes and reports a player's playback rate. FilePlayer
// values may capability-assert this optional interface, but only positive rates
// are supported for streaming playback.
type VariableRatePlayer interface {
	SetRate(rate float32) error
	Rate() (float32, error)
}

// RateModulatedPlayer accepts a signal controlling native playback rate.
type RateModulatedPlayer interface{ SetRateModulator(Signal) error }

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

// MicrophonePermission is the observable result of a microphone access request.
type MicrophonePermission uint8

const (
	MicrophonePermissionPending MicrophonePermission = iota
	MicrophonePermissionDenied
	MicrophonePermissionGranted
)

// MicrophoneSource selects the physical recording input.
type MicrophoneSource uint8

const (
	MicrophoneSourceAutomatic MicrophoneSource = iota
	MicrophoneSourceInternal
	MicrophoneSourceHeadset
)

// MicrophoneSamples is valid only for the duration of its recording callback.
// CopyTo copies at most len(destination) mono PCM samples and never retains it.
type MicrophoneSamples interface {
	Len() int
	CopyTo(destination []int16) (int, error)
}

// MicrophoneRecorder owns the currently installed native recording callback.
type MicrophoneRecorder interface {
	Source() MicrophoneSource
	Stop() error
	Close() error
}

// Microphones provides permission-gated microphone recording. Starting a new
// recorder stops and closes the previous recorder owned by this capability.
type Microphones interface {
	RequestMicrophoneAccess(purpose string, callback func(MicrophonePermission)) (MicrophonePermission, error)
	StartMicrophoneRecording(source MicrophoneSource, callback func(MicrophoneSamples) bool) (MicrophoneRecorder, error)
}

// AudioChannel owns one native routing node and its source attachments. Closing
// a source detaches it; closing the channel detaches every source. Neither
// operation closes the other owned object.
type AudioChannel interface {
	AddSource(source AudioSource) error
	RemoveSource(source AudioSource) error
	AddEffect(effect AudioEffect) error
	RemoveEffect(effect AudioEffect) error
	SetVolume(volume float32) error
	Volume() (float32, error)
	SetPan(pan float32) error
	SetVolumeModulator(Signal) error
	SetPanModulator(Signal) error
	Close() error
}

// AudioEffect is an explicitly owned channel processor. Attaching an effect
// never transfers ownership; closing either endpoint detaches the graph edge.
type AudioEffect interface {
	SetMix(level float32) error
	SetMixModulator(signal Signal) error
	Close() error
}

// FilterType selects the response of a two-pole filter.
type FilterType uint8

const (
	FilterLowPass FilterType = iota
	FilterHighPass
	FilterBandPass
	FilterNotch
	FilterPEQ
	FilterLowShelf
	FilterHighShelf
)

type TwoPoleFilter interface {
	AudioEffect
	SetFrequency(float32) error
	SetFrequencyModulator(Signal) error
	SetGain(float32) error
	SetResonance(float32) error
	SetResonanceModulator(Signal) error
}

type OnePoleFilter interface {
	AudioEffect
	SetParameter(float32) error
	SetParameterModulator(Signal) error
}

type BitCrusher interface {
	AudioEffect
	SetExponential(bool) error
	SetDepth(float32) error
	SetDepthModulator(Signal) error
	SetDownsampling(float32) error
	SetDownsamplingModulator(Signal) error
}

type RingModulator interface {
	AudioEffect
	SetFrequency(float32) error
	SetFrequencyModulator(Signal) error
}

type Overdrive interface {
	AudioEffect
	SetGain(float32) error
	SetLimit(float32) error
	SetLimitModulator(Signal) error
	SetOffset(float32) error
	SetOffsetModulator(Signal) error
}

// DelayLine is an owned effect whose taps are independently owned sources.
// Close detaches and closes all taps before freeing the delay.
type DelayLine interface {
	AudioEffect
	SetLength(frames int) error
	SetFeedback(float32) error
	AddTap(delayFrames int) (DelayTap, error)
}

type DelayTap interface {
	AudioSource
	SetDelay(frames int) error
	SetDelayModulator(Signal) error
	SetChannelsFlipped(bool) error
	Close() error
}

// AudioEffects creates owned native processors.
type AudioEffects interface {
	NewTwoPoleFilter(FilterType) (TwoPoleFilter, error)
	NewOnePoleFilter() (OnePoleFilter, error)
	NewBitCrusher() (BitCrusher, error)
	NewRingModulator() (RingModulator, error)
	NewDelayLine(lengthFrames int, stereo bool) (DelayLine, error)
	NewOverdrive() (Overdrive, error)
}

// AudioChannels creates explicitly owned audio routing channels. Games should
// capability-assert this optional interface from Context.
type AudioChannels interface {
	NewAudioChannel() (AudioChannel, error)
}

// AudioOutputState reports the currently connected physical audio outputs.
// HeadsetMicrophone is meaningful only when Headphones is true.
type AudioOutputState struct {
	Headphones        bool
	HeadsetMicrophone bool
}

// AudioOutputs controls the device audio outputs and exposes the default
// channel. Output activation is hardware state: Simulator implementations may
// accept it without producing an observable host-routing change.
type AudioOutputs interface {
	DefaultAudioChannel() (AudioChannel, error)
	AudioOutputState() (AudioOutputState, error)
	SetAudioOutputsActive(headphones, speaker bool) error
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
	SetArpeggiation(steps []float32) error
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
	SetCurvature(amount float32) error
	SetVelocitySensitivity(amount float32) error
	SetRateScaling(scaling float32, startNote, endNote uint8) error
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
	SetEnvelopeCurvature(amount float32) error
	SetEnvelopeVelocitySensitivity(amount float32) error
	SetEnvelopeRateScaling(scaling float32, startNote, endNote uint8) error
	SetTranspose(semitones float32) error
	SetFrequencyModulator(signal Signal) error
	SetAmplitudeModulator(signal Signal) error
	SetWavetable(sample AudioSample, log2Size, columns, rows int) error
	SetParameter(parameter int, value float32) error
	SetParameterModulator(parameter int, signal Signal) error
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

// Instrument owns a native voice bank but not the Synth values assigned to it.
type Instrument interface {
	AudioSource
	AddVoice(synth Synth, rangeStart, rangeEnd uint8, transpose float32) error
	SetPitchBend(float32) error
	SetPitchBendRange(float32) error
	SetTranspose(float32) error
	NoteOff(note uint8, when uint32) error
	AllNotesOff(when uint32) error
	SetVolume(left, right float32) error
	Volume() (left, right float32, err error)
	ActiveVoiceCount() (int, error)
	Close() error
}

// SequenceTrack owns note and control events. Its instrument attachment does
// not transfer ownership and is detached when either value closes.
type SequenceTrack interface {
	SetInstrument(Instrument) error
	AddNote(step, length uint32, note uint8, velocity float32) error
	RemoveNote(step uint32, note uint8) error
	ClearNotes() error
	AddControlEvent(controller, step int, value float32, interpolate bool) error
	RemoveControlEvent(controller, step int) error
	ClearControlEvents() error
	SetMuted(bool) error
	Length() (uint32, error)
	Instrument() (Instrument, error)
	ControlSignalCount() (int, error)
	ControlSignal(index int) (ControlSignal, error)
	SignalForController(controller int, create bool) (ControlSignal, error)
	Polyphony() (int, error)
	ActiveVoiceCount() (int, error)
	NoteIndexAtStep(step uint32) (int, error)
	NoteAt(index int) (SequenceNote, bool, error)
	Close() error
}

// SequenceNote is an immutable snapshot of one note event.
type SequenceNote struct {
	Step, Length uint32
	Note         uint8
	Velocity     float32
}

// Sequence owns playback state and MIDI-loaded children, but not tracks
// attached programmatically. LoadMIDI must precede programmatic SetTrack calls.
type Sequence interface {
	LoadMIDI(path string) error
	SetTempo(stepsPerSecond float32) error
	Tempo() (float32, error)
	SetLoops(start, end, count int) error
	SetTrack(index uint, track SequenceTrack) error
	Play(callback func()) error
	Stop() error
	IsPlaying() (bool, error)
	Time() (uint32, error)
	SetTime(uint32) error
	Length() (uint32, error)
	TrackCount() (uint, error)
	Track(index uint) (SequenceTrack, error)
	CurrentStep() (step, timeOffset int, err error)
	AllNotesOff() error
	Close() error
}

// Sequencers creates explicitly owned dynamic-music graph nodes.
type Sequencers interface {
	NewInstrument() (Instrument, error)
	NewSequenceTrack() (SequenceTrack, error)
	NewSequence() (Sequence, error)
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

// StreamingPlayerControls configures file reload, buffering, looping, and
// underrun behavior without widening the base FilePlayer contract.
type StreamingPlayerControls interface {
	Load(path string) error
	SetBufferLength(seconds float32) error
	SetLoopRange(start, end float32) error
	DidUnderrun() (bool, error)
	SetStopOnUnderrun(bool) error
}

// Audio loads the two accepted audio use cases.
type Audio interface {
	LoadSoundEffect(path string) (SoundEffect, error)
	LoadFilePlayer(path string) (FilePlayer, error)
}
