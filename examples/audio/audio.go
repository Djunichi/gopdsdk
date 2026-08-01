// Package audio exercises the two portable P2.4 audio use cases.
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
	effect playdate.SoundEffect
	music  playdate.FilePlayer
	closed bool
}

// New creates the P2.4 audio acceptance game.
func New() playdate.Game { return &game{} }

func (g *game) Init(context playdate.Context) error {
	effect, err := context.LoadSoundEffect(effectAsset)
	if err != nil {
		return err
	}
	g.effect = effect
	music, err := context.LoadFilePlayer(musicAsset)
	if err != nil {
		return errors.Join(err, effect.Close())
	}
	g.music = music
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
		if err := g.effect.Play(); err != nil {
			return false, err
		}
	}
	if input.Pressed.Has(playdate.ButtonB) {
		state, err := g.music.State()
		if err != nil {
			return false, err
		}
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
	context.Clear()
	context.DrawText("P2.4 audio", 12, 8)
	context.DrawText("A: repeat sound effect", 12, 32)
	context.DrawText("B: play/stop music", 12, 54)
	context.DrawText("SFX: "+stateName(effectState), 12, 86)
	context.DrawText("Music: "+stateName(musicState), 12, 108)
	return true, nil
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
