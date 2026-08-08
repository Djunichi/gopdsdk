package runtime

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/Djunichi/gopdsdk/playdate"
)

func TestOwnedFontMeasurementAndClose(t *testing.T) {
	freed := uintptr(0)
	font := NewOwnedFont(17, FontDriver{
		TextWidth: func(handle uintptr, text string) int { return int(handle) + len(text) },
		Height:    func(handle uintptr) int { return int(handle) },
		Free:      func(handle uintptr) { freed = handle },
	})
	if width, err := font.TextWidth("HUD"); err != nil || width != 20 {
		t.Fatalf("TextWidth = %d, %v", width, err)
	}
	if height, err := font.Height(); err != nil || height != 17 {
		t.Fatalf("Height = %d, %v", height, err)
	}
	if err := font.Close(); err != nil || freed != 17 {
		t.Fatalf("Close = %v, freed %d", err, freed)
	}
	if _, err := font.TextWidth("HUD"); !errors.Is(err, playdate.ErrFontClosed) {
		t.Fatalf("closed TextWidth = %v", err)
	}
	if err := font.Close(); !errors.Is(err, playdate.ErrFontClosed) {
		t.Fatalf("second Close = %v", err)
	}
}

func TestFontHandleRejectsForeignFont(t *testing.T) {
	if _, err := FontHandle(foreignFont{}); !errors.Is(err, playdate.ErrFontInvalid) {
		t.Fatalf("FontHandle = %v", err)
	}
}

type foreignFont struct{}

func (foreignFont) TextWidth(string) (int, error) { return 0, nil }
func (foreignFont) Height() (int, error)          { return 0, nil }
func (foreignFont) Close() error                  { return nil }

type testContext struct{}

func (testContext) Clear()                                                             {}
func (testContext) DrawText(string, int, int)                                          {}
func (testContext) LoadFont(string) (playdate.Font, error)                             { return foreignFont{}, nil }
func (testContext) DrawTextFont(playdate.Font, string, int, int) error                 { return nil }
func (testContext) CurrentTimeMilliseconds() uint32                                    { return 0 }
func (testContext) Input() playdate.Input                                              { return playdate.Input{} }
func (testContext) LoadBitmap(string) (playdate.Bitmap, error)                         { return nil, nil }
func (testContext) LoadBitmapTable(string) (playdate.BitmapTable, error)               { return nil, nil }
func (testContext) NewBitmap(int, int) (playdate.Bitmap, error)                        { return nil, nil }
func (testContext) DrawBitmap(playdate.Bitmap, int, int) error                         { return nil }
func (testContext) DrawScaledBitmap(playdate.Bitmap, int, int, float32, float32) error { return nil }
func (testContext) NewSprite() (playdate.Sprite, error)                                { return nil, nil }
func (testContext) QuerySpritesAtPoint(float32, float32) []playdate.Sprite             { return nil }
func (testContext) QuerySpritesInRect(playdate.Rect) []playdate.Sprite                 { return nil }
func (testContext) QueryOverlappingSprites(playdate.Sprite) ([]playdate.Sprite, error) {
	return nil, nil
}
func (testContext) UpdateAndDrawSprites()                                {}
func (testContext) LoadSoundEffect(string) (playdate.SoundEffect, error) { return nil, nil }
func (testContext) LoadFilePlayer(string) (playdate.FilePlayer, error)   { return nil, nil }

type launcherContext struct {
	testContext
	exited bool
}

func (context *launcherContext) ExitToLauncher() { context.exited = true }

type graphicsCapabilityContext struct {
	testContext
	draws, framebuffers, offscreen, transformed, stencils int
}

func (c *graphicsCapabilityContext) DrawLine(int, int, int, int, int, playdate.Paint) error {
	c.draws++
	return nil
}
func (*graphicsCapabilityContext) DrawRect(int, int, int, int, playdate.Paint) error { return nil }
func (*graphicsCapabilityContext) FillRect(int, int, int, int, playdate.Paint) error { return nil }
func (*graphicsCapabilityContext) DrawEllipse(int, int, int, int, int, float32, float32, playdate.Paint) error {
	return nil
}
func (*graphicsCapabilityContext) FillEllipse(int, int, int, int, float32, float32, playdate.Paint) error {
	return nil
}
func (*graphicsCapabilityContext) DrawTriangle(int, int, int, int, int, int, int, playdate.Paint) error {
	return nil
}
func (*graphicsCapabilityContext) FillTriangle(int, int, int, int, int, int, playdate.Paint) error {
	return nil
}
func (*graphicsCapabilityContext) SetClipRect(int, int, int, int) error { return nil }
func (*graphicsCapabilityContext) ClearClipRect()                       {}
func (*graphicsCapabilityContext) SetDrawOffset(int, int)               {}
func (*graphicsCapabilityContext) SetDrawMode(playdate.DrawMode) error  { return nil }
func (c *graphicsCapabilityContext) WithFramebuffer(callback func(playdate.Framebuffer) error) error {
	c.framebuffers++
	return callback(nil)
}
func (c *graphicsCapabilityContext) DrawInto(_ playdate.Bitmap, callback func() error) error {
	c.offscreen++
	return callback()
}
func (c *graphicsCapabilityContext) DrawRotatedBitmap(playdate.Bitmap, int, int, float32, float32, float32, float32, float32) error {
	c.transformed++
	return nil
}
func (c *graphicsCapabilityContext) WithStencil(_ playdate.Bitmap, _ bool, callback func() error) error {
	c.stencils++
	return callback()
}

type graphicsCapabilityGame struct{}

func (graphicsCapabilityGame) Init(context playdate.Context) error {
	if _, ok := context.(playdate.PrimitiveGraphics); !ok {
		return playdate.ErrGraphicsUnavailable
	}
	if _, ok := context.(playdate.GraphicsState); !ok {
		return playdate.ErrGraphicsUnavailable
	}
	if _, ok := context.(playdate.FramebufferGraphics); !ok {
		return playdate.ErrGraphicsUnavailable
	}
	if _, ok := context.(playdate.OffscreenGraphics); !ok {
		return playdate.ErrGraphicsUnavailable
	}
	if _, ok := context.(playdate.BitmapCompositor); !ok {
		return playdate.ErrGraphicsUnavailable
	}
	return nil
}
func (graphicsCapabilityGame) Update(context playdate.Context) (bool, error) {
	paint, _ := playdate.SolidPaint(playdate.ColorBlack)
	if err := context.(playdate.PrimitiveGraphics).DrawLine(0, 0, 1, 1, 1, paint); err != nil {
		return false, err
	}
	if err := context.(playdate.FramebufferGraphics).WithFramebuffer(func(playdate.Framebuffer) error { return nil }); err != nil {
		return false, err
	}
	if err := context.(playdate.BitmapCompositor).DrawRotatedBitmap(nil, 10, 20, 45, .5, .5, 1, 1); err != nil {
		return false, err
	}
	if err := context.(playdate.BitmapCompositor).WithStencil(nil, false, func() error {
		return context.(playdate.BitmapCompositor).WithStencil(nil, false, func() error { return nil })
	}); !errors.Is(err, playdate.ErrGraphicsStencilActive) {
		return false, err
	}
	if err := context.(playdate.BitmapCompositor).WithStencil(nil, false, func() error { return nil }); err != nil {
		return false, err
	}
	return true, context.(playdate.OffscreenGraphics).DrawInto(nil, func() error { return nil })
}

func TestApplicationForwardsOptionalGraphicsCapabilities(t *testing.T) {
	context := &graphicsCapabilityContext{}
	application, err := NewApplication(graphicsCapabilityGame{}, context, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if refresh, err := application.Update(RawInput{}); err != nil || refresh != 1 {
		t.Fatalf("Update() = %v, %v", refresh, err)
	}
	if context.draws != 1 {
		t.Fatalf("draws = %d", context.draws)
	}
	if context.framebuffers != 1 || context.offscreen != 1 {
		t.Fatalf("framebuffers/offscreen = %d/%d", context.framebuffers, context.offscreen)
	}
	if context.transformed != 1 || context.stencils != 2 {
		t.Fatalf("transformed/stencils = %d/%d", context.transformed, context.stencils)
	}
}

func TestAudioOwnershipStateAndValidation(t *testing.T) {
	playing := false
	paused := false
	freed := 0
	left, right := float32(1), float32(1)
	driver := AudioDriver{
		Play:      func(uintptr) bool { playing = true; return true },
		Stop:      func(uintptr) { playing = false },
		SetVolume: func(_ uintptr, l, r float32) { left, right = l, r },
		Volume:    func(uintptr) (float32, float32) { return left, right },
		IsPlaying: func(uintptr) bool { return playing },
		Pause:     func(_ uintptr, value bool) { paused = value },
		Free:      func(uintptr) { freed++ },
	}
	player := NewSoundEffect(17, driver)
	if err := player.SetVolume(.25, .75); err != nil {
		t.Fatal(err)
	}
	if l, r, err := player.Volume(); err != nil || l != .25 || r != .75 {
		t.Fatalf("Volume() = %v, %v, %v", l, r, err)
	}
	if err := player.Play(); err != nil {
		t.Fatal(err)
	}
	if state, _ := player.State(); state != playdate.PlaybackPlaying {
		t.Fatalf("state = %v", state)
	}
	if err := player.Pause(); err != nil || !paused {
		t.Fatalf("Pause() = %v, paused %v", err, paused)
	}
	if state, _ := player.State(); state != playdate.PlaybackPaused {
		t.Fatalf("state = %v", state)
	}
	if err := player.Resume(); err != nil || paused {
		t.Fatalf("Resume() = %v, paused %v", err, paused)
	}
	if err := player.Stop(); err != nil {
		t.Fatal(err)
	}
	if state, _ := player.State(); state != playdate.PlaybackStopped {
		t.Fatalf("state = %v", state)
	}
	if err := player.Play(); err != nil {
		t.Fatal(err)
	}
	if err := player.Close(); err != nil || freed != 1 || playing {
		t.Fatalf("Close() = %v, freed %d, playing %v", err, freed, playing)
	}
	if err := player.Play(); !errors.Is(err, playdate.ErrAudioClosed) {
		t.Fatalf("closed Play() = %v", err)
	}
	if err := ValidateAudioVolume(-1, 1); !errors.Is(err, playdate.ErrAudioVolume) {
		t.Fatalf("volume error = %v", err)
	}
}

func TestAudioPlayFailure(t *testing.T) {
	player := NewFilePlayer(1, AudioDriver{Play: func(uintptr) bool { return false }})
	if err := player.Play(); !errors.Is(err, playdate.ErrAudioPlay) {
		t.Fatalf("Play() = %v", err)
	}
}

func TestFilePlayerVariableRate(t *testing.T) {
	rate := float32(1)
	player := NewFilePlayer(1, AudioDriver{
		SetRate: func(_ uintptr, value float32) { rate = value },
		Rate:    func(uintptr) float32 { return rate },
	})
	variable, ok := player.(playdate.VariableRatePlayer)
	if !ok {
		t.Fatal("FilePlayer does not expose VariableRatePlayer")
	}
	if err := variable.SetRate(.75); err != nil {
		t.Fatal(err)
	}
	if value, err := variable.Rate(); err != nil || value != .75 {
		t.Fatalf("Rate() = %v, %v", value, err)
	}
	if err := variable.SetRate(-1); !errors.Is(err, playdate.ErrAudioReverseUnsupported) {
		t.Fatalf("reverse file rate = %v", err)
	}
	if rate != .75 {
		t.Fatalf("native rate changed after rejected reverse = %v", rate)
	}
}

func TestAudioCompletionAndFadeCallbacks(t *testing.T) {
	var finishID, fadeID uint32
	player := NewFilePlayer(5, AudioDriver{
		SetFinishCallback: func(_ uintptr, callback uint32) { finishID = callback },
		FadeVolume: func(_ uintptr, left, right float32, frames, callback uint32) {
			if left != .25 || right != .75 || frames != 4410 {
				t.Fatalf("fade = %v, %v, %d", left, right, frames)
			}
			fadeID = callback
		},
		Stop: func(uintptr) {}, Free: func(uintptr) {},
	})
	completed, faded := 0, 0
	if err := player.(playdate.CompletionPlayer).SetFinishCallback(func() { completed++ }); err != nil {
		t.Fatal(err)
	}
	if finishID == 0 {
		t.Fatal("finish callback was not registered")
	}
	InvokeAudioCallback(finishID, false)
	InvokeAudioCallback(finishID, false)
	DrainAudioCallbacks()
	if completed != 2 {
		t.Fatalf("completed = %d", completed)
	}
	oldFinishID := finishID
	if err := player.(playdate.CompletionPlayer).SetFinishCallback(func() { completed += 10 }); err != nil {
		t.Fatal(err)
	}
	InvokeAudioCallback(oldFinishID, false)
	InvokeAudioCallback(finishID, false)
	DrainAudioCallbacks()
	if completed != 12 {
		t.Fatalf("completed after replacement = %d", completed)
	}
	if err := player.(playdate.FadingPlayer).FadeVolume(.25, .75, 4410, func() { faded++ }); err != nil {
		t.Fatal(err)
	}
	InvokeAudioCallback(fadeID, true)
	InvokeAudioCallback(fadeID, true)
	DrainAudioCallbacks()
	if faded != 1 {
		t.Fatalf("faded = %d", faded)
	}
	if err := player.(playdate.FadingPlayer).FadeVolume(1, 1, 2147483648, nil); !errors.Is(err, playdate.ErrAudioFade) {
		t.Fatalf("large fade = %v", err)
	}
	if err := player.Close(); err != nil {
		t.Fatal(err)
	}
	if finishID != 0 {
		t.Fatalf("native finish callback after Close = %d", finishID)
	}
}

func TestSamplePlayerDoesNotExposeFadingPlayer(t *testing.T) {
	player := NewSamplePlayer(1, AudioDriver{})
	if _, ok := player.(playdate.FadingPlayer); ok {
		t.Fatal("SamplePlayer exposes streaming-only fades")
	}
	if _, ok := player.(playdate.CompletionPlayer); !ok {
		t.Fatal("SamplePlayer does not expose completion callbacks")
	}
}

type audioClockContext struct{ testContext }

func (audioClockContext) CurrentAudioTime() (uint32, error) { return 12345, nil }

func TestApplicationForwardsAudioClock(t *testing.T) {
	application, err := NewApplication(testGame{init: func(context playdate.Context) error {
		value, clockErr := context.(playdate.AudioClock).CurrentAudioTime()
		if value != 12345 || clockErr != nil {
			t.Fatalf("CurrentAudioTime = %d, %v", value, clockErr)
		}
		return nil
	}}, audioClockContext{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
}

func TestSamplePlayerControls(t *testing.T) {
	playingRepeat := 0
	playingRate := float32(0)
	offset := float32(0)
	rate := float32(1)
	player := NewSamplePlayer(9, AudioDriver{
		PlayRepeated: func(_ uintptr, repeat int, value float32) bool {
			playingRepeat, playingRate = repeat, value
			return true
		},
		Length:    func(uintptr) float32 { return 2.5 },
		SetOffset: func(_ uintptr, value float32) { offset = value }, Offset: func(uintptr) float32 { return offset },
		SetRate: func(_ uintptr, value float32) { rate = value }, Rate: func(uintptr) float32 { return rate },
		Stop: func(uintptr) {}, Free: func(uintptr) {},
	})
	if err := player.PlayRepeated(3, .5); err != nil || playingRepeat != 3 || playingRate != .5 {
		t.Fatalf("PlayRepeated() = %v, %d, %v", err, playingRepeat, playingRate)
	}
	if length, err := player.Length(); err != nil || length != 2.5 {
		t.Fatalf("Length() = %v, %v", length, err)
	}
	if err := player.SetOffset(1.25); err != nil {
		t.Fatal(err)
	}
	if value, err := player.Offset(); err != nil || value != 1.25 {
		t.Fatalf("Offset() = %v, %v", value, err)
	}
	if err := player.SetRate(-1); err != nil {
		t.Fatal(err)
	}
	if value, err := player.Rate(); err != nil || value != -1 {
		t.Fatalf("Rate() = %v, %v", value, err)
	}
	if err := player.PlayRepeated(-1, 1); !errors.Is(err, playdate.ErrAudioRepeat) {
		t.Fatalf("negative repeat = %v", err)
	}
	if int(^uint(0)>>63) == 1 {
		if err := player.PlayRepeated(int(int64(2147483648)), 1); !errors.Is(err, playdate.ErrAudioRepeat) {
			t.Fatalf("large repeat = %v", err)
		}
	}
	if err := player.SetRate(0); !errors.Is(err, playdate.ErrAudioRate) {
		t.Fatalf("zero rate = %v", err)
	}
	if err := player.SetOffset(-1); !errors.Is(err, playdate.ErrAudioOffset) {
		t.Fatalf("negative offset = %v", err)
	}
}

type samplePlayerContext struct {
	testContext
	loaded string
}

func (context *samplePlayerContext) LoadSamplePlayer(path string) (playdate.SamplePlayer, error) {
	context.loaded = path
	return NewSamplePlayer(1, AudioDriver{Stop: func(uintptr) {}, Free: func(uintptr) {}}), nil
}

func TestApplicationForwardsSamplePlayers(t *testing.T) {
	context := &samplePlayerContext{}
	application, err := NewApplication(testGame{init: func(got playdate.Context) error {
		samples, ok := got.(playdate.SamplePlayers)
		if !ok {
			return errors.New("sample players not forwarded")
		}
		_, loadErr := samples.LoadSamplePlayer("audio/hit")
		return loadErr
	}}, context, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if context.loaded != "audio/hit" {
		t.Fatalf("loaded = %q", context.loaded)
	}
}

func TestAudioChannelOwnership(t *testing.T) {
	var added, removed, freed int
	source := NewSamplePlayer(7, AudioDriver{Source: func(uintptr) uintptr { return 70 }, Stop: func(uintptr) {}, Free: func(uintptr) {}})
	channel := NewAudioChannel(9, AudioChannelDriver{
		AddSource:    func(channel, source uintptr) bool { added++; return channel == 9 && source == 70 },
		RemoveSource: func(channel, source uintptr) bool { removed++; return channel == 9 && source == 70 },
		SetVolume:    func(uintptr, float32) {}, Volume: func(uintptr) float32 { return .5 }, SetPan: func(uintptr, float32) {},
		Remove: func(uintptr) bool { return true }, Free: func(uintptr) { freed++ },
	})
	if err := channel.AddSource(source); err != nil {
		t.Fatal(err)
	}
	if err := channel.AddSource(source); err != nil || added != 2 {
		t.Fatalf("refresh AddSource = %v, %d", err, added)
	}
	if err := source.Close(); err != nil || removed != 1 {
		t.Fatalf("source Close = %v, removed %d", err, removed)
	}
	if err := channel.Close(); err != nil || freed != 1 {
		t.Fatalf("channel Close = %v, freed %d", err, freed)
	}
	if err := channel.SetPan(0); !errors.Is(err, playdate.ErrAudioChannelClosed) {
		t.Fatalf("closed SetPan = %v", err)
	}
}

func TestSynthSignalOwnership(t *testing.T) {
	var frequency, amplitude uintptr
	var signalFreed, synthFreed int
	signalDriver := SignalDriver{Value: func(uintptr) float32 { return .25 }, SetScale: func(uintptr, float32) {}, SetOffset: func(uintptr, float32) {}, Free: func(uintptr) { signalFreed++ }}
	lfo := NewLFO(3, LFODriver{Signal: signalDriver, SetRate: func(uintptr, float32) {}, SetPhase: func(uintptr, float32) {}, SetCenter: func(uintptr, float32) {}, SetDepth: func(uintptr, float32) {}, SetRetrigger: func(uintptr, bool) {}})
	synth := NewSynth(4, SynthDriver{
		Audio:       AudioDriver{Source: func(handle uintptr) uintptr { return handle }, Stop: func(uintptr) {}, Free: func(uintptr) { synthFreed++ }},
		SetWaveform: func(uintptr, playdate.Waveform) {}, SetEnvelope: func(uintptr, float32, float32, float32, float32) {}, SetTranspose: func(uintptr, float32) {},
		SetFrequencyModulator: func(_ uintptr, signal uintptr) { frequency = signal }, SetAmplitudeModulator: func(_ uintptr, signal uintptr) { amplitude = signal },
		PlayMIDINote: func(uintptr, float32, float32, float32, uint32) {}, NoteOff: func(uintptr, uint32) {},
	})
	if err := synth.SetFrequencyModulator(lfo); err != nil || frequency != 3 {
		t.Fatalf("SetFrequencyModulator = %v, %d", err, frequency)
	}
	if err := synth.SetAmplitudeModulator(lfo); err != nil || amplitude != 3 {
		t.Fatalf("SetAmplitudeModulator = %v, %d", err, amplitude)
	}
	if err := lfo.Close(); err != nil || frequency != 0 || amplitude != 0 || signalFreed != 1 {
		t.Fatalf("LFO Close = %v, %d, %d, %d", err, frequency, amplitude, signalFreed)
	}
	if err := synth.PlayMIDINote(60, 1, .5, 100); err != nil {
		t.Fatal(err)
	}
	if err := synth.PlayMIDINote(60, 2, .5, 100); !errors.Is(err, playdate.ErrAudioParameter) {
		t.Fatalf("velocity = %v", err)
	}
	if err := synth.Close(); err != nil || synthFreed != 1 {
		t.Fatalf("Synth Close = %v, %d", err, synthFreed)
	}
}

func TestAudioChannelRoutesSynthSource(t *testing.T) {
	synth := NewSynth(4, SynthDriver{Audio: AudioDriver{Source: func(handle uintptr) uintptr { return handle }}})
	var routed uintptr
	adds := 0
	channel := NewAudioChannel(9, AudioChannelDriver{AddSource: func(_, source uintptr) bool { routed = source; adds++; return true }})
	if err := channel.AddSource(synth); err != nil || routed != 4 {
		t.Fatalf("AddSource(synth) = %v, routed %d", err, routed)
	}
	if err := channel.AddSource(synth); err != nil || adds != 2 {
		t.Fatalf("AddSource(synth) refresh = %v, calls %d", err, adds)
	}
}

func TestAudioChannelMovesSourceFromPreviousChannel(t *testing.T) {
	synth := NewSynth(4, SynthDriver{Audio: AudioDriver{Source: func(handle uintptr) uintptr { return handle }}})
	var removedFrom, addedTo uintptr
	driver := AudioChannelDriver{
		AddSource:    func(channel, _ uintptr) bool { addedTo = channel; return true },
		RemoveSource: func(channel, _ uintptr) bool { removedFrom = channel; return true },
	}
	first := NewAudioChannel(8, driver)
	second := NewAudioChannel(9, driver)
	if err := first.AddSource(synth); err != nil {
		t.Fatal(err)
	}
	if err := second.AddSource(synth); err != nil {
		t.Fatal(err)
	}
	if removedFrom != 8 || addedTo != 9 {
		t.Fatalf("move route removed from %d, added to %d", removedFrom, addedTo)
	}
	if err := first.RemoveSource(synth); err != nil || removedFrom != 8 {
		t.Fatalf("stale first route = %v, removed from %d", err, removedFrom)
	}
	if err := second.RemoveSource(synth); err != nil || removedFrom != 9 {
		t.Fatalf("active second route = %v, removed from %d", err, removedFrom)
	}
}

type musicContext struct{ testContext }

func TestAudioEffectGraphDetachesBothDirections(t *testing.T) {
	var added, removed uintptr
	channel := NewAudioChannel(9, AudioChannelDriver{
		AddEffect:    func(_, effect uintptr) bool { added = effect; return true },
		RemoveEffect: func(_, effect uintptr) bool { removed = effect; return true },
		Remove:       func(uintptr) bool { return true }, Free: func(uintptr) {},
	})
	var modulator uintptr
	freed := 0
	effect := NewRingModulator(7, RingModulatorDriver{
		Effect:       EffectDriver{SetMix: func(uintptr, float32) {}, SetMixModulator: func(uintptr, uintptr) {}, Free: func(uintptr) { freed++ }},
		SetFrequency: func(uintptr, float32) {}, SetFrequencyModulator: func(_, signal uintptr) { modulator = signal },
	})
	signal := NewLFO(3, LFODriver{Signal: SignalDriver{Value: func(uintptr) float32 { return 0 }, SetScale: func(uintptr, float32) {}, SetOffset: func(uintptr, float32) {}, Free: func(uintptr) {}}})
	if err := channel.AddEffect(effect); err != nil || added != 7 {
		t.Fatalf("AddEffect = %v, %d", err, added)
	}
	if err := effect.SetFrequencyModulator(signal); err != nil || modulator != 3 {
		t.Fatalf("SetFrequencyModulator = %v, %d", err, modulator)
	}
	if err := signal.Close(); err != nil || modulator != 0 {
		t.Fatalf("signal Close = %v, modulator %d", err, modulator)
	}
	if err := effect.Close(); err != nil || removed != 7 || freed != 1 {
		t.Fatalf("effect Close = %v, removed %d, freed %d", err, removed, freed)
	}
}

func TestAudioChannelRoutesDelayTap(t *testing.T) {
	var routed uintptr
	channel := NewAudioChannel(9, AudioChannelDriver{AddSource: func(_, source uintptr) bool { routed = source; return true }})
	delay := NewDelayLine(7, DelayLineDriver{AddTap: func(uintptr, int) uintptr { return 5 }, Tap: AudioDriver{Source: func(h uintptr) uintptr { return h }}})
	tap, err := delay.AddTap(100)
	if err != nil {
		t.Fatal(err)
	}
	if err = channel.AddSource(tap); err != nil || routed != 5 {
		t.Fatalf("AddSource(delay tap) = %v, routed %d", err, routed)
	}
}

func TestSequenceOwnershipAndCallbackLifetime(t *testing.T) {
	var trackInstrument, routedTrack uintptr
	var callback uint32
	trackCount := uint(0)
	freed := [3]int{}
	synth := NewSynth(1, SynthDriver{Audio: AudioDriver{Source: func(h uintptr) uintptr { return h }, Stop: func(uintptr) {}, Free: func(uintptr) { freed[0]++ }}}).(*synth)
	instrument := NewInstrument(2, InstrumentDriver{AddVoice: func(_, _ uintptr, _, _ uint8, _ float32) bool { return true }, Free: func(uintptr) { freed[1]++ }}).(*instrument)
	track := NewSequenceTrack(3, TrackDriver{SetInstrument: func(_, value uintptr) { trackInstrument = value }, Free: func(uintptr) { freed[2]++ }}).(*sequenceTrack)
	sequence := NewSequence(4, SequenceDriver{TrackCount: func(uintptr) uint { return trackCount }, AddTrack: func(uintptr) uintptr { trackCount++; return 99 }, SetTrack: func(_ uintptr, _ uint, value uintptr) { routedTrack = value }, Play: func(_ uintptr, value uint32) { callback = value }, Stop: func(uintptr) {}, Free: func(uintptr) {}}).(*sequence)
	if err := instrument.AddVoice(synth, 0, 127, 0); err != nil {
		t.Fatal(err)
	}
	if err := instrument.AddVoice(synth, 0, 127, 0); !errors.Is(err, playdate.ErrAudioRoute) {
		t.Fatalf("duplicate instrument voice = %v", err)
	}
	if err := NewAudioChannel(8, AudioChannelDriver{}).AddSource(synth); !errors.Is(err, playdate.ErrAudioRoute) {
		t.Fatalf("instrument synth channel route = %v", err)
	}
	if err := synth.Close(); !errors.Is(err, playdate.ErrAudioRoute) {
		t.Fatalf("attached synth Close = %v", err)
	}
	if err := track.SetInstrument(instrument); err != nil || trackInstrument != 2 {
		t.Fatalf("SetInstrument = %v, %d", err, trackInstrument)
	}
	if err := sequence.SetTrack(0, track); err != nil || routedTrack != 3 {
		t.Fatalf("SetTrack = %v, %d", err, routedTrack)
	}
	if trackCount != 1 {
		t.Fatalf("native track slots = %d", trackCount)
	}
	called := 0
	if err := sequence.Play(func() { called++ }); err != nil || callback == 0 {
		t.Fatalf("Play = %v, callback %d", err, callback)
	}
	InvokeAudioCallback(callback, true)
	DrainAudioCallbacks()
	if called != 1 {
		t.Fatalf("completion calls = %d", called)
	}
	if err := instrument.Close(); err != nil || trackInstrument != 0 {
		t.Fatalf("instrument Close = %v, track instrument %d", err, trackInstrument)
	}
	if err := synth.Close(); err != nil || freed[0] != 1 {
		t.Fatalf("detached synth Close = %v, frees %d", err, freed[0])
	}
	if err := track.Close(); err != nil || routedTrack != 0 || freed[2] != 1 {
		t.Fatalf("track Close = %v, sequence track %d, frees %d", err, routedTrack, freed[2])
	}
}

func TestInstrumentRejectsChannelRoutedSynth(t *testing.T) {
	synth := NewSynth(1, SynthDriver{Audio: AudioDriver{Source: func(h uintptr) uintptr { return h }}})
	channel := NewAudioChannel(2, AudioChannelDriver{AddSource: func(uintptr, uintptr) bool { return true }})
	if err := channel.AddSource(synth); err != nil {
		t.Fatal(err)
	}
	instrument := NewInstrument(3, InstrumentDriver{AddVoice: func(uintptr, uintptr, uint8, uint8, float32) bool { t.Fatal("native addVoice called"); return true }})
	if err := instrument.AddVoice(synth, 0, 127, 0); !errors.Is(err, playdate.ErrAudioRoute) {
		t.Fatalf("AddVoice routed synth = %v", err)
	}
}

func (musicContext) NewAudioChannel() (playdate.AudioChannel, error) {
	return NewAudioChannel(1, AudioChannelDriver{}), nil
}
func (musicContext) NewSynth(playdate.Waveform) (playdate.Synth, error) {
	return NewSynth(1, SynthDriver{}), nil
}
func (musicContext) NewLFO(playdate.LFOType) (playdate.LFO, error) {
	return NewLFO(1, LFODriver{}), nil
}
func (musicContext) NewEnvelope(float32, float32, float32, float32) (playdate.Envelope, error) {
	return NewEnvelope(1, EnvelopeDriver{}), nil
}
func (musicContext) NewControlSignal() (playdate.ControlSignal, error) {
	return NewControlSignal(1, ControlSignalDriver{}), nil
}
func (musicContext) NewInstrument() (playdate.Instrument, error)       { return nil, nil }
func (musicContext) NewSequenceTrack() (playdate.SequenceTrack, error) { return nil, nil }
func (musicContext) NewSequence() (playdate.Sequence, error)           { return nil, nil }
func (musicContext) NewTwoPoleFilter(playdate.FilterType) (playdate.TwoPoleFilter, error) {
	return nil, nil
}
func (musicContext) NewBitCrusher() (playdate.BitCrusher, error)        { return nil, nil }
func (musicContext) NewRingModulator() (playdate.RingModulator, error)  { return nil, nil }
func (musicContext) NewDelayLine(int, bool) (playdate.DelayLine, error) { return nil, nil }
func (musicContext) NewOverdrive() (playdate.Overdrive, error)          { return nil, nil }

func TestApplicationForwardsMusicGraph(t *testing.T) {
	application, err := NewApplication(testGame{init: func(context playdate.Context) error {
		if _, ok := context.(playdate.AudioChannels); !ok {
			return errors.New("audio channels not forwarded")
		}
		if _, ok := context.(playdate.Synthesizers); !ok {
			return errors.New("synthesizers not forwarded")
		}
		sequencers, ok := context.(playdate.Sequencers)
		if !ok {
			return errors.New("sequencers not forwarded")
		}
		if _, err := sequencers.NewSequence(); err != nil {
			return err
		}
		effects, ok := context.(playdate.AudioEffects)
		if !ok {
			return errors.New("audio effects not forwarded")
		}
		if _, err := effects.NewOverdrive(); err != nil {
			return err
		}
		return nil
	}}, musicContext{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
}

func TestBitmapOwnershipLifecycle(t *testing.T) {
	var freed []uintptr
	var fills []playdate.Color
	driver := BitmapDriver{
		Dimensions: func(uintptr) (int, int) { return 400, 240 },
		Fill:       func(_ uintptr, color playdate.Color) { fills = append(fills, color) },
		Free:       func(handle uintptr) { freed = append(freed, handle) },
	}
	owned := NewOwnedBitmap(7, driver)
	if width, err := owned.Width(); err != nil || width != 400 {
		t.Fatalf("Width() = %d, %v", width, err)
	}
	if height, err := owned.Height(); err != nil || height != 240 {
		t.Fatalf("Height() = %d, %v", height, err)
	}
	if err := owned.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := owned.Fill(playdate.ColorBlack); err != nil {
		t.Fatal(err)
	}
	if err := owned.Fill(playdate.Color(99)); !errors.Is(err, playdate.ErrBitmapColor) {
		t.Fatalf("invalid Fill() error = %v", err)
	}
	if err := owned.Close(); err != nil {
		t.Fatal(err)
	}
	if len(freed) != 1 || freed[0] != 7 {
		t.Fatalf("freed = %v", freed)
	}
	if len(fills) != 2 || fills[0] != playdate.ColorClear || fills[1] != playdate.ColorBlack {
		t.Fatalf("fills = %v", fills)
	}
	for name, operation := range map[string]func() error{
		"double close":           owned.Close,
		"fill after close":       func() error { return owned.Fill(playdate.ColorWhite) },
		"dimensions after close": func() error { _, err := owned.Width(); return err },
	} {
		if err := operation(); !errors.Is(err, playdate.ErrBitmapClosed) {
			t.Fatalf("%s error = %v", name, err)
		}
	}
	borrowed := NewBorrowedBitmap(8, driver)
	if err := borrowed.Close(); !errors.Is(err, playdate.ErrBitmapBorrowed) {
		t.Fatalf("borrowed Close() = %v", err)
	}
	if _, err := BitmapHandle(borrowed); err != nil {
		t.Fatal(err)
	}
	if _, err := OwnedBitmapHandle(borrowed); !errors.Is(err, playdate.ErrBitmapBorrowed) {
		t.Fatalf("borrowed offscreen handle = %v", err)
	}
	owned = NewOwnedBitmap(9, driver)
	if handle, err := OwnedBitmapHandle(owned); err != nil || handle != 9 {
		t.Fatalf("owned offscreen handle = %d, %v", handle, err)
	}
}

func TestBitmapTableOwnershipAndBorrowedFrames(t *testing.T) {
	freed := uintptr(0)
	table := NewOwnedBitmapTable(11, BitmapTableDriver{Frame: func(_ uintptr, index int) uintptr {
		if index == 2 {
			return 22
		}
		return 0
	}, Free: func(handle uintptr) { freed = handle }}, BitmapDriver{})
	frame, err := table.Frame(2)
	if err != nil {
		t.Fatal(err)
	}
	if err := frame.Close(); !errors.Is(err, playdate.ErrBitmapBorrowed) {
		t.Fatalf("frame close = %v", err)
	}
	if _, err := table.Frame(1); !errors.Is(err, playdate.ErrBitmapFrameRange) {
		t.Fatalf("range error = %v", err)
	}
	if err := table.Close(); err != nil {
		t.Fatal(err)
	}
	if freed != 11 {
		t.Fatalf("freed = %d", freed)
	}
	if _, err := frame.Width(); !errors.Is(err, playdate.ErrBitmapClosed) {
		t.Fatalf("frame after table close = %v", err)
	}
	if _, err := table.Frame(2); !errors.Is(err, playdate.ErrBitmapTableClosed) {
		t.Fatalf("closed error = %v", err)
	}
}

func TestBitmapArgumentValidation(t *testing.T) {
	if err := ValidateBitmapSize(0, 1); !errors.Is(err, playdate.ErrBitmapSize) {
		t.Fatalf("size error = %v", err)
	}
	if err := ValidateBitmapScale(1, 0); !errors.Is(err, playdate.ErrBitmapScale) {
		t.Fatalf("scale error = %v", err)
	}
	if err := ValidateBitmapScale(1, *(*float32)(unsafe.Pointer(&[]uint32{0x7fc00000}[0]))); !errors.Is(err, playdate.ErrBitmapScale) {
		t.Fatalf("NaN scale error = %v", err)
	}
	if err := ValidateBitmapTransform(*(*float32)(unsafe.Pointer(&[]uint32{0x7f800000}[0])), .5, .5, 1, 1); !errors.Is(err, playdate.ErrGraphicsGeometry) {
		t.Fatalf("infinite rotation error = %v", err)
	}
	narrow := NewOwnedBitmap(7, BitmapDriver{Dimensions: func(uintptr) (int, int) { return 31, 8 }})
	if _, err := ValidateStencil(narrow, true); !errors.Is(err, playdate.ErrGraphicsStencilWidth) {
		t.Fatalf("tiled stencil width error = %v", err)
	}
	wide := NewOwnedBitmap(8, BitmapDriver{Dimensions: func(uintptr) (int, int) { return 32, 8 }})
	if handle, err := ValidateStencil(wide, true); err != nil || handle != 8 {
		t.Fatalf("valid tiled stencil = %d, %v", handle, err)
	}
}

func TestPrimitiveAndDrawModeValidation(t *testing.T) {
	if err := ValidatePrimitiveGeometry(0, 1, 1, 0, 360); !errors.Is(err, playdate.ErrGraphicsGeometry) {
		t.Fatalf("width error = %v", err)
	}
	if err := ValidatePrimitiveGeometry(1, 1, 1, 0, 360); err != nil {
		t.Fatalf("valid geometry = %v", err)
	}
	if err := ValidateDrawMode(playdate.DrawMode(99)); !errors.Is(err, playdate.ErrGraphicsDrawMode) {
		t.Fatalf("draw mode error = %v", err)
	}
}

func TestSpriteDisplayListOwnershipLifecycle(t *testing.T) {
	var operations []string
	driver := SpriteDriver{
		SetBitmap:  func(sprite, bitmap uintptr) { operations = append(operations, "bitmap") },
		MoveTo:     func(uintptr, float32, float32) { operations = append(operations, "position") },
		MoveBy:     func(uintptr, float32, float32) { operations = append(operations, "move") },
		SetVisible: func(uintptr, bool) { operations = append(operations, "visible") },
		SetZIndex:  func(uintptr, int) { operations = append(operations, "z") },
		MarkDirty:  func(uintptr) { operations = append(operations, "dirty") },
		MarkDirtyRect: func(uintptr, playdate.Rect) {
			operations = append(operations, "dirtyRect")
		},
		Add:    func(uintptr) { operations = append(operations, "add") },
		Remove: func(uintptr) { operations = append(operations, "remove") },
		Free:   func(uintptr) { operations = append(operations, "free") },
	}
	sprite := NewOwnedSprite(9, driver)
	bitmap := NewOwnedBitmap(7, BitmapDriver{})
	if err := sprite.SetBitmap(bitmap); err != nil {
		t.Fatal(err)
	}
	if err := sprite.SetPosition(1, 2); err != nil {
		t.Fatal(err)
	}
	if err := sprite.MoveBy(3, 4); err != nil {
		t.Fatal(err)
	}
	if err := sprite.SetVisible(true); err != nil {
		t.Fatal(err)
	}
	if err := sprite.SetZIndex(5); err != nil {
		t.Fatal(err)
	}
	if err := sprite.MarkDirty(); err != nil {
		t.Fatal(err)
	}
	if err := sprite.MarkDirtyRect(playdate.Rect{Width: 2, Height: 3}); err != nil {
		t.Fatal(err)
	}
	if err := sprite.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sprite.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sprite.Close(); err != nil {
		t.Fatal(err)
	}
	want := []string{"bitmap", "position", "move", "visible", "z", "dirty", "dirtyRect", "add", "remove", "free"}
	if len(operations) != len(want) {
		t.Fatalf("operations = %v, want %v", operations, want)
	}
	for index := range want {
		if operations[index] != want[index] {
			t.Fatalf("operations = %v, want %v", operations, want)
		}
	}
	if err := sprite.Close(); !errors.Is(err, playdate.ErrSpriteClosed) {
		t.Fatalf("double Close() = %v", err)
	}
}

func TestDisplayAndDirtyRegionValidation(t *testing.T) {
	var displayCalls int
	display := NewDisplay(DisplayDriver{
		SetRefreshRate: func(float32) { displayCalls++ }, SetInverted: func(bool) { displayCalls++ },
		SetScale: func(uint) { displayCalls++ }, SetMosaic: func(uint, uint) { displayCalls++ },
		SetFlipped: func(bool, bool) { displayCalls++ }, SetOffset: func(int, int) { displayCalls++ },
	})
	if err := display.SetRefreshRate(50); err != nil {
		t.Fatal(err)
	}
	if err := display.SetScale(4); err != nil {
		t.Fatal(err)
	}
	if err := display.SetMosaic(3, 0); err != nil {
		t.Fatal(err)
	}
	display.SetInverted(true)
	display.SetFlipped(true, false)
	display.SetOffset(1, -1)
	if displayCalls != 6 {
		t.Fatalf("display calls = %d", displayCalls)
	}
	if err := display.SetRefreshRate(51); !errors.Is(err, playdate.ErrDisplayRefreshRate) {
		t.Fatalf("refresh error = %v", err)
	}
	if err := display.SetScale(3); !errors.Is(err, playdate.ErrDisplayScale) {
		t.Fatalf("scale error = %v", err)
	}
	if err := display.SetMosaic(0, 4); !errors.Is(err, playdate.ErrDisplayMosaic) {
		t.Fatalf("mosaic error = %v", err)
	}
	if err := ValidateScreenDirtyRect(0, 0, 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := ValidateScreenDirtyRect(0, 0, 0, 1); !errors.Is(err, playdate.ErrSpriteDirtyRect) {
		t.Fatalf("screen rect error = %v", err)
	}
	if err := ValidateSpriteRect(playdate.Rect{Width: 1, Height: 1}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSpriteRect(playdate.Rect{Width: -1, Height: 1}); !errors.Is(err, playdate.ErrSpriteDirtyRect) {
		t.Fatalf("sprite rect error = %v", err)
	}
}

func TestVideoPlayerOwnershipValidationAndErrors(t *testing.T) {
	var freed int
	bitmap := NewOwnedBitmap(7, BitmapDriver{Dimensions: func(uintptr) (int, int) { return 80, 60 }, Free: func(uintptr) {}})
	player := NewVideoPlayer(9, VideoDriver{
		Info: func(uintptr) playdate.VideoInfo {
			return playdate.VideoInfo{Width: 80, Height: 60, FrameRate: 20, FrameCount: 3}
		},
		SetContext:       func(player, context uintptr) (bool, string) { return player == 9 && context == 7, "bad context" },
		UseScreenContext: func(uintptr) {},
		RenderFrame:      func(_ uintptr, frame int) (bool, string) { return frame != 1, "decode failed" },
		Free:             func(uintptr) { freed++ },
	})
	if info, err := player.Info(); err != nil || info.FrameCount != 3 {
		t.Fatalf("Info() = %+v, %v", info, err)
	}
	if err := player.SetContext(bitmap); err != nil {
		t.Fatal(err)
	}
	if err := player.RenderFrame(-1); !errors.Is(err, playdate.ErrVideoFrame) {
		t.Fatalf("negative frame error = %v", err)
	}
	if err := player.RenderFrame(1); err == nil || err.Error() != "render frame video failed: decode failed" {
		t.Fatalf("decoder error = %v", err)
	}
	if err := bitmap.Close(); err != nil {
		t.Fatal(err)
	}
	if err := player.RenderFrame(0); !errors.Is(err, playdate.ErrBitmapClosed) {
		t.Fatalf("closed context error = %v", err)
	}
	if err := player.UseScreenContext(); err != nil {
		t.Fatal(err)
	}
	if err := player.RenderFrame(0); err != nil {
		t.Fatal(err)
	}
	if err := player.Close(); err != nil || freed != 1 {
		t.Fatalf("Close() = %v, frees %d", err, freed)
	}
	if err := player.Close(); !errors.Is(err, playdate.ErrVideoClosed) {
		t.Fatalf("second Close() = %v", err)
	}
}

type videoContext struct {
	testContext
	player playdate.VideoPlayer
}

func (c videoContext) LoadVideo(path string) (playdate.VideoPlayer, error) {
	if path != "intro.pdv" {
		return nil, errors.New("wrong path")
	}
	return c.player, nil
}

type videoCapabilityGame struct{ got playdate.VideoPlayer }

func (g *videoCapabilityGame) Init(context playdate.Context) error {
	player, err := context.(playdate.Videos).LoadVideo("intro.pdv")
	g.got = player
	return err
}
func (*videoCapabilityGame) Update(playdate.Context) (bool, error) { return false, nil }

func TestNewApplicationForwardsVideoCapability(t *testing.T) {
	player := NewVideoPlayer(1, VideoDriver{})
	game := &videoCapabilityGame{}
	application, err := NewApplication(game, videoContext{player: player}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if game.got != player {
		t.Fatal("video player was not forwarded")
	}
}

type displayCapabilityContext struct {
	testContext
	calls int
}

func (c *displayCapabilityContext) SetRefreshRate(float32) error          { c.calls++; return nil }
func (*displayCapabilityContext) SetInverted(bool)                        {}
func (*displayCapabilityContext) SetScale(uint) error                     { return nil }
func (*displayCapabilityContext) SetMosaic(uint, uint) error              { return nil }
func (*displayCapabilityContext) SetFlipped(bool, bool)                   {}
func (*displayCapabilityContext) SetOffset(int, int)                      {}
func (c *displayCapabilityContext) SetAlwaysRedraw(bool)                  { c.calls++ }
func (c *displayCapabilityContext) AddDirtyRect(int, int, int, int) error { c.calls++; return nil }

type displayCapabilityGame struct{}

func (displayCapabilityGame) Init(context playdate.Context) error {
	if err := context.(playdate.Display).SetRefreshRate(30); err != nil {
		return err
	}
	context.(playdate.SpriteRedraw).SetAlwaysRedraw(false)
	return context.(playdate.SpriteRedraw).AddDirtyRect(1, 2, 3, 4)
}
func (displayCapabilityGame) Update(playdate.Context) (bool, error) { return false, nil }

func TestApplicationForwardsDisplayAndSpriteRedrawCapabilities(t *testing.T) {
	context := &displayCapabilityContext{}
	application, err := NewApplication(displayCapabilityGame{}, context, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if context.calls != 3 {
		t.Fatalf("forwarded calls = %d", context.calls)
	}
}

func TestSpriteCollisionBridgeAndBorrowedResponse(t *testing.T) {
	var operations []string
	driver := SpriteDriver{
		SetCollideRect:   func(uintptr, playdate.Rect) { operations = append(operations, "rect") },
		ClearCollideRect: func(uintptr) { operations = append(operations, "clear") },
		SetTag:           func(uintptr, uint8) { operations = append(operations, "tag") },
		MoveWithCollisions: func(uintptr, float32, float32) (float32, float32, []NativeCollision) {
			return 8, 9, []NativeCollision{{Other: 12, ResponseType: playdate.CollisionOverlap, Time: .5}}
		},
	}
	sprite := NewOwnedSprite(11, driver)
	if err := sprite.SetCollideRect(playdate.Rect{Width: 16, Height: 16}); err != nil {
		t.Fatal(err)
	}
	if err := sprite.SetTag(2); err != nil {
		t.Fatal(err)
	}
	result, err := sprite.MoveWithCollisions(10, 10)
	if err != nil {
		t.Fatal(err)
	}
	if result.ActualX != 8 || result.ActualY != 9 || len(result.Collisions) != 1 || result.Collisions[0].ResponseType != playdate.CollisionOverlap {
		t.Fatalf("result = %+v", result)
	}
	if err := result.Collisions[0].Other.Close(); !errors.Is(err, playdate.ErrSpriteBorrowed) {
		t.Fatalf("borrowed Close() = %v", err)
	}
	if err := sprite.ClearCollideRect(); err != nil {
		t.Fatal(err)
	}
	want := []string{"rect", "tag", "clear"}
	if len(operations) != len(want) {
		t.Fatalf("operations = %v", operations)
	}
	for index := range want {
		if operations[index] != want[index] {
			t.Fatalf("operations = %v", operations)
		}
	}
}

type testGame struct {
	init   func(playdate.Context) error
	update func(playdate.Context) (bool, error)
}

func (g testGame) Init(context playdate.Context) error { return g.init(context) }
func (g testGame) Update(context playdate.Context) (bool, error) {
	return g.update(context)
}

func TestABIFixedWidthTypes(t *testing.T) {
	if got := unsafe.Sizeof(Event(0)); got != 4 {
		t.Fatalf("sizeof(Event) = %d, want 4", got)
	}
}

func TestNewRequiresCallbacks(t *testing.T) {
	if _, err := New(Callbacks{}); !errors.Is(err, ErrInitRequired) {
		t.Fatalf("New() error = %v, want ErrInitRequired", err)
	}
	if _, err := New(Callbacks{Init: func() error { return nil }}); !errors.Is(err, ErrUpdateRequired) {
		t.Fatalf("New() error = %v, want ErrUpdateRequired", err)
	}
}

func TestRuntimeLifecycle(t *testing.T) {
	initCalls := 0
	updateCalls := 0
	runtime, err := New(Callbacks{
		Init: func() error {
			initCalls++
			return nil
		},
		Update: func(playdate.Input) (bool, error) {
			updateCalls++
			return true, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Update() before init error = %v, want ErrNotInitialized", err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	refresh, err := runtime.Update(RawInput{})
	if err != nil {
		t.Fatal(err)
	}
	if refresh != 1 || initCalls != 1 || updateCalls != 1 {
		t.Fatalf("refresh/init/update = %d/%d/%d, want 1/1/1", refresh, initCalls, updateCalls)
	}
	if err := runtime.Handle(EventInit, 0); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("duplicate Handle(EventInit) error = %v, want ErrAlreadyInitialized", err)
	}
}

func TestRuntimeRefreshAndErrors(t *testing.T) {
	updateErr := errors.New("update failed")
	tests := []struct {
		name        string
		update      func(playdate.Input) (bool, error)
		wantRefresh int32
		wantErr     error
	}{
		{"no refresh", func(playdate.Input) (bool, error) { return false, nil }, 0, nil},
		{"update error", func(playdate.Input) (bool, error) { return true, updateErr }, 0, updateErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := New(Callbacks{Init: func() error { return nil }, Update: test.update})
			if err != nil {
				t.Fatal(err)
			}
			if err := runtime.Handle(EventInit, 0); err != nil {
				t.Fatal(err)
			}
			refresh, err := runtime.Update(RawInput{})
			if refresh != test.wantRefresh || !errors.Is(err, test.wantErr) {
				t.Fatalf("Update() = %d, %v; want %d, %v", refresh, err, test.wantRefresh, test.wantErr)
			}
		})
	}
}

func TestRuntimeFailedInitializationCannotUpdate(t *testing.T) {
	initErr := errors.New("init failed")
	runtime, err := New(Callbacks{
		Init:   func() error { return initErr },
		Update: func(playdate.Input) (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); !errors.Is(err, initErr) {
		t.Fatalf("Handle() error = %v, want init error", err)
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrFailed) {
		t.Fatalf("Update() error = %v, want ErrFailed", err)
	}
}

func TestApplicationIsCommonLifecycleEntryPoint(t *testing.T) {
	var calls []string
	context := &launcherContext{}
	application, err := NewApplication(testGame{
		init: func(got playdate.Context) error {
			calls = append(calls, "init")
			fonts, ok := got.(playdate.FontGraphics)
			if !ok {
				return errors.New("font graphics not forwarded")
			}
			if _, fontErr := fonts.LoadFont("fonts/test"); fontErr != nil {
				return fontErr
			}
			launcher, ok := got.(playdate.Launcher)
			if !ok {
				return errors.New("launcher not forwarded")
			}
			launcher.ExitToLauncher()
			return nil
		},
		update: func(got playdate.Context) (bool, error) {
			calls = append(calls, "update")
			return true, nil
		},
	}, context, func() { calls = append(calls, "before-init") })
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	refresh, err := application.Update(RawInput{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := refresh, int32(1); got != want {
		t.Fatalf("Update() refresh = %d, want %d", got, want)
	}
	wantCalls := []string{"before-init", "init", "update"}
	if len(calls) != len(wantCalls) {
		t.Fatalf("calls = %v, want %v", calls, wantCalls)
	}
	for index := range wantCalls {
		if calls[index] != wantCalls[index] {
			t.Fatalf("calls = %v, want %v", calls, wantCalls)
		}
	}
	if !context.exited {
		t.Fatal("ExitToLauncher was not forwarded")
	}
}

func TestLifecycleEventsPreservePlatformOrder(t *testing.T) {
	var got []playdate.LifecycleEvent
	runtime, err := New(Callbacks{
		Init: func() error { return nil },
		Lifecycle: func(event playdate.LifecycleEvent) error {
			got = append(got, event)
			return nil
		},
		Update: func(playdate.Input) (bool, error) { return false, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	events := []Event{EventPause, EventLock, EventLowPower, EventUnlock, EventResume, EventTerminate}
	for _, event := range events {
		if err := runtime.Handle(event, 0); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrTerminated) {
		t.Fatalf("Update() after terminate error = %v, want ErrTerminated", err)
	}
	want := []playdate.LifecycleEvent{playdate.LifecyclePause, playdate.LifecycleLock, playdate.LifecycleLowPower, playdate.LifecycleUnlock, playdate.LifecycleResume, playdate.LifecycleTerminate}
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}

func TestLifecycleErrorIsFailStop(t *testing.T) {
	wantErr := errors.New("pause failed")
	runtime, err := New(Callbacks{
		Init:      func() error { return nil },
		Lifecycle: func(playdate.LifecycleEvent) error { return wantErr },
		Update:    func(playdate.Input) (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventPause, 0); !errors.Is(err, wantErr) {
		t.Fatalf("pause error = %v, want %v", err, wantErr)
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrFailed) {
		t.Fatalf("Update() error = %v, want ErrFailed", err)
	}
}

func TestInputTransitionsBetweenFrames(t *testing.T) {
	var snapshots []playdate.Input
	runtime, err := New(Callbacks{
		Init: func() error { return nil },
		Update: func(input playdate.Input) (bool, error) {
			snapshots = append(snapshots, input)
			return false, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	frames := []RawInput{
		{Buttons: playdate.ButtonA, CrankAngle: 10, CrankDelta: 2, CrankDocked: true, DeltaSeconds: 0.016},
		{Buttons: playdate.ButtonA | playdate.ButtonLeft, CrankAngle: 15, CrankDelta: 5, CrankDocked: false, DeltaSeconds: 0.017},
		{Buttons: playdate.ButtonLeft, CrankAngle: 15, CrankDocked: true, DeltaSeconds: 0.018},
	}
	for _, frame := range frames {
		if _, err := runtime.Update(frame); err != nil {
			t.Fatal(err)
		}
	}
	if got := snapshots[0]; got.Pressed != playdate.ButtonA || got.Held != 0 || got.Released != 0 || got.CrankDockedThisFrame || got.CrankUndocked {
		t.Fatalf("first snapshot transitions = %+v", got)
	}
	if got := snapshots[1]; got.Pressed != playdate.ButtonLeft || got.Held != playdate.ButtonA || got.Released != 0 || !got.CrankUndocked || got.DeltaSeconds != 0.017 {
		t.Fatalf("second snapshot transitions = %+v", got)
	}
	if got := snapshots[2]; got.Pressed != 0 || got.Held != playdate.ButtonLeft || got.Released != playdate.ButtonA || !got.CrankDockedThisFrame {
		t.Fatalf("third snapshot transitions = %+v", got)
	}
}
