package audio

import (
	"errors"
	"testing"

	"github.com/Djunichi/gopdsdk/playdate"
)

type player struct {
	state          playdate.PlaybackState
	closed, paused bool
	plays          int
	volume         float32
	rate, offset   float32
	repeat         int
	finish         func()
	fade           func()
	fadeFrames     uint32
}

func (p *player) Play() error                            { p.state = playdate.PlaybackPlaying; p.plays++; return nil }
func (p *player) Stop() error                            { p.state = playdate.PlaybackStopped; p.paused = false; return nil }
func (p *player) SetVolume(left, _ float32) error        { p.volume = left; return nil }
func (p *player) Volume() (float32, float32, error)      { return p.volume, p.volume, nil }
func (p *player) State() (playdate.PlaybackState, error) { return p.state, nil }
func (p *player) Pause() error {
	if p.state == playdate.PlaybackPlaying {
		p.state = playdate.PlaybackPaused
		p.paused = true
	}
	return nil
}
func (p *player) Resume() error {
	if p.paused {
		p.state = playdate.PlaybackPlaying
		p.paused = false
	}
	return nil
}
func (p *player) Close() error { p.closed = true; p.state = playdate.PlaybackStopped; return nil }
func (p *player) PlayRepeated(repeat int, rate float32) error {
	p.repeat, p.rate, p.state = repeat, rate, playdate.PlaybackPlaying
	p.plays++
	return nil
}
func (*player) Length() (float32, error)                  { return 2.5, nil }
func (p *player) SetOffset(value float32) error           { p.offset = value; return nil }
func (p *player) Offset() (float32, error)                { return p.offset, nil }
func (p *player) SetRate(value float32) error             { p.rate = value; return nil }
func (p *player) Rate() (float32, error)                  { return p.rate, nil }
func (p *player) SetFinishCallback(callback func()) error { p.finish = callback; return nil }
func (p *player) FadeVolume(left, _ float32, frames uint32, callback func()) error {
	p.volume, p.fadeFrames, p.fade = left, frames, callback
	return nil
}

type context struct {
	effect   *player
	music    *player
	input    playdate.Input
	musicErr error
}

func (*context) CurrentTimeMilliseconds() uint32                                    { return 0 }
func (*context) CurrentAudioTime() (uint32, error)                                  { return 44100, nil }
func (c *context) Input() playdate.Input                                            { return c.input }
func (*context) Clear()                                                             {}
func (*context) DrawText(string, int, int)                                          {}
func (*context) LoadBitmap(string) (playdate.Bitmap, error)                         { return nil, nil }
func (*context) LoadBitmapTable(string) (playdate.BitmapTable, error)               { return nil, nil }
func (*context) NewBitmap(int, int) (playdate.Bitmap, error)                        { return nil, nil }
func (*context) DrawBitmap(playdate.Bitmap, int, int) error                         { return nil }
func (*context) DrawScaledBitmap(playdate.Bitmap, int, int, float32, float32) error { return nil }
func (*context) NewSprite() (playdate.Sprite, error)                                { return nil, nil }
func (*context) QuerySpritesAtPoint(float32, float32) []playdate.Sprite             { return nil }
func (*context) QuerySpritesInRect(playdate.Rect) []playdate.Sprite                 { return nil }
func (*context) QueryOverlappingSprites(playdate.Sprite) ([]playdate.Sprite, error) { return nil, nil }
func (*context) UpdateAndDrawSprites()                                              {}
func (c *context) LoadSoundEffect(path string) (playdate.SoundEffect, error) {
	if path != effectAsset {
		panic(path)
	}
	c.effect = &player{}
	return c.effect, nil
}
func (c *context) LoadSamplePlayer(path string) (playdate.SamplePlayer, error) {
	if path != effectAsset {
		panic(path)
	}
	c.effect = &player{rate: 1}
	return c.effect, nil
}
func (c *context) LoadFilePlayer(path string) (playdate.FilePlayer, error) {
	if path != musicAsset {
		panic(path)
	}
	if c.musicErr != nil {
		return nil, c.musicErr
	}
	c.music = &player{}
	return c.music, nil
}

func TestRepeatedEffectMusicAndLifecycle(t *testing.T) {
	c := &context{}
	g := New().(*game)
	if err := g.Init(c); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		c.input.Pressed = playdate.ButtonA
		if _, err := g.Update(c); err != nil {
			t.Fatal(err)
		}
	}
	if c.effect.plays != 2 {
		t.Fatalf("effect plays = %d", c.effect.plays)
	}
	if c.effect.repeat != 3 || c.effect.rate != 1 {
		t.Fatalf("sample repeat/rate = %d/%v", c.effect.repeat, c.effect.rate)
	}
	c.input.Pressed = playdate.ButtonRight
	if _, err := g.Update(c); err != nil {
		t.Fatal(err)
	}
	if c.effect.rate != 1.25 {
		t.Fatalf("sample rate = %v", c.effect.rate)
	}
	c.input.Pressed = playdate.ButtonUp
	if _, err := g.Update(c); err != nil {
		t.Fatal(err)
	}
	if c.music.rate != 1.25 {
		t.Fatalf("music rate = %v", c.music.rate)
	}
	c.input.Pressed = playdate.ButtonB
	if _, err := g.Update(c); err != nil {
		t.Fatal(err)
	}
	c.input.Pressed = playdate.ButtonB
	if _, err := g.Update(c); err != nil {
		t.Fatal(err)
	}
	if c.music.fadeFrames != 22050 || c.music.fade == nil {
		t.Fatalf("fade frames = %d, callback present = %t", c.music.fadeFrames, c.music.fade != nil)
	}
	c.music.fade()
	c.effect.finish()
	if g.fadeFinished != 1 || g.sampleFinished != 1 || g.audioTime != 44100 {
		t.Fatalf("callbacks/time = %d/%d/%d", g.fadeFinished, g.sampleFinished, g.audioTime)
	}
	if err := g.HandleLifecycle(c, playdate.LifecyclePause); err != nil {
		t.Fatal(err)
	}
	if c.effect.state != playdate.PlaybackPaused || c.music.state != playdate.PlaybackPaused {
		t.Fatalf("paused states = %v/%v", c.effect.state, c.music.state)
	}
	if err := g.HandleLifecycle(c, playdate.LifecycleResume); err != nil {
		t.Fatal(err)
	}
	if c.effect.state != playdate.PlaybackPlaying || c.music.state != playdate.PlaybackPlaying {
		t.Fatalf("resumed states = %v/%v", c.effect.state, c.music.state)
	}
	if err := g.HandleLifecycle(c, playdate.LifecycleTerminate); err != nil {
		t.Fatal(err)
	}
	if !c.effect.closed || !c.music.closed {
		t.Fatal("players were not closed")
	}
}

func TestInitializationRollback(t *testing.T) {
	want := errors.New("music load")
	c := &context{musicErr: want}
	err := New().Init(c)
	if !errors.Is(err, want) || c.effect == nil || !c.effect.closed {
		t.Fatalf("Init() = %v, effect = %+v", err, c.effect)
	}
}
