// Package audio exercises P5.1 sample playback and P2.4 streaming audio.
package audio

import (
	"errors"

	"github.com/Djunichi/gopdsdk/playdate"
)

const (
	effectAsset = "audio/effect"
	musicAsset  = "audio/music"
)

type game struct {
	effect          playdate.SamplePlayer
	music           playdate.FilePlayer
	musicRatePlayer playdate.VariableRatePlayer
	sampleRate      float32
	musicRate       float32
	length          float32
	effectState     playdate.PlaybackState
	musicState      playdate.PlaybackState
	dirty           bool
	closed          bool
}

// New creates the P5.1 audio acceptance game.
func New() playdate.Game { return &game{} }

func (g *game) Init(context playdate.Context) error {
	samples, ok := context.(playdate.SamplePlayers)
	if !ok {
		return playdate.ErrAudioUnavailable
	}
	effect, err := samples.LoadSamplePlayer(effectAsset)
	if err != nil {
		return err
	}
	g.effect = effect
	g.sampleRate = 1
	g.musicRate = 1
	g.dirty = true
	g.length, err = effect.Length()
	if err != nil {
		return errors.Join(err, effect.Close())
	}
	music, err := context.LoadFilePlayer(musicAsset)
	if err != nil {
		return errors.Join(err, effect.Close())
	}
	g.music = music
	musicRatePlayer, ok := music.(playdate.VariableRatePlayer)
	if !ok {
		return errors.Join(playdate.ErrAudioUnavailable, g.close())
	}
	g.musicRatePlayer = musicRatePlayer
	if err = effect.SetVolume(.8, .8); err != nil {
		return errors.Join(err, g.close())
	}
	if err = music.SetVolume(.35, .35); err != nil {
		return errors.Join(err, g.close())
	}
	return nil
}

func (g *game) Update(context playdate.Context) (bool, error) {
	input := context.Input()
	if input.Pressed.Has(playdate.ButtonA) {
		offset := float32(0)
		if g.sampleRate < 0 {
			offset = g.length
		}
		if err := g.effect.SetOffset(offset); err != nil {
			return false, err
		}
		if err := g.effect.PlayRepeated(3, g.sampleRate); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if input.Pressed.Has(playdate.ButtonRight) && g.sampleRate < 2 {
		g.sampleRate = nextRate(g.sampleRate, .25)
		if err := g.effect.SetRate(g.sampleRate); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if input.Pressed.Has(playdate.ButtonLeft) && g.sampleRate > -2 {
		g.sampleRate = nextRate(g.sampleRate, -.25)
		if err := g.effect.SetRate(g.sampleRate); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if input.Pressed.Has(playdate.ButtonUp) && g.musicRate < 2 {
		g.musicRate = nextRate(g.musicRate, .25)
		if err := g.musicRatePlayer.SetRate(g.musicRate); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if input.Pressed.Has(playdate.ButtonDown) && g.musicRate > .25 {
		g.musicRate = nextRate(g.musicRate, -.25)
		if err := g.musicRatePlayer.SetRate(g.musicRate); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if input.Pressed.Has(playdate.ButtonB) {
		state, err := g.music.State()
		if err != nil {
			return false, err
		}
		g.dirty = true
		if state == playdate.PlaybackStopped {
			err = g.music.Play()
		} else {
			err = g.music.Stop()
		}
		if err != nil {
			return false, err
		}
	}
	effectState, err := g.effect.State()
	if err != nil {
		return false, err
	}
	musicState, err := g.music.State()
	if err != nil {
		return false, err
	}
	if effectState != g.effectState || musicState != g.musicState {
		g.effectState, g.musicState = effectState, musicState
		g.dirty = true
	}
	if !g.dirty {
		return false, nil
	}
	g.dirty = false
	context.Clear()
	context.DrawText("P5.1 sample playback", 12, 8)
	context.DrawText("A: play sample x3", 12, 32)
	context.DrawText("B: play/stop music", 12, 54)
	context.DrawText("Left/Right: sample rate", 12, 76)
	context.DrawText("Up/Down: music rate", 12, 98)
	context.DrawText("SFX: "+stateName(effectState)+" "+rateName(g.sampleRate), 12, 130)
	context.DrawText("Music: "+stateName(musicState)+" "+rateName(g.musicRate), 12, 152)
	return true, nil
}

func nextRate(rate, delta float32) float32 {
	rate += delta
	if rate == 0 {
		if delta < 0 {
			return -.25
		}
		return .25
	}
	return rate
}

func rateName(rate float32) string {
	switch rate {
	case -2:
		return "-2.00x"
	case -1.75:
		return "-1.75x"
	case -1.5:
		return "-1.50x"
	case -1.25:
		return "-1.25x"
	case -1:
		return "-1.00x"
	case -.75:
		return "-0.75x"
	case -.5:
		return "-0.50x"
	case -.25:
		return "-0.25x"
	case .25:
		return "0.25x"
	case .5:
		return "0.50x"
	case .75:
		return "0.75x"
	case 1:
		return "1.00x"
	case 1.25:
		return "1.25x"
	case 1.5:
		return "1.50x"
	case 1.75:
		return "1.75x"
	case 2:
		return "2.00x"
	default:
		return "invalid"
	}
}

func (g *game) HandleLifecycle(_ playdate.Context, event playdate.LifecycleEvent) error {
	if event == playdate.LifecyclePause || event == playdate.LifecycleLock {
		return errors.Join(g.effect.Pause(), g.music.Pause())
	}
	if event == playdate.LifecycleResume || event == playdate.LifecycleUnlock {
		return errors.Join(g.effect.Resume(), g.music.Resume())
	}
	if event == playdate.LifecycleTerminate {
		return g.close()
	}
	return nil
}

func (g *game) close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	var musicErr, effectErr error
	if g.music != nil {
		musicErr = g.music.Close()
	}
	if g.effect != nil {
		effectErr = g.effect.Close()
	}
	return errors.Join(musicErr, effectErr)
}

func stateName(state playdate.PlaybackState) string {
	if state == playdate.PlaybackPlaying {
		return "playing"
	}
	if state == playdate.PlaybackPaused {
		return "paused"
	}
	return "stopped"
}
