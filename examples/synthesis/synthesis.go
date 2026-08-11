// Package synthesis demonstrates P9.3 envelope shaping and native synthesis.
package synthesis

import "github.com/Djunichi/gopdsdk/playdate"

type game struct {
	synth     playdate.Synth
	clock     playdate.AudioClock
	curvature float32
	dirty     bool
}

// New creates the synthesis acceptance scene.
func New() playdate.Game { return &game{curvature: 0, dirty: true} }

func (g *game) Init(context playdate.Context) error {
	synths, ok := context.(playdate.Synthesizers)
	if !ok {
		return playdate.ErrAudioUnavailable
	}
	clock, ok := context.(playdate.AudioClock)
	if !ok {
		return playdate.ErrAudioUnavailable
	}
	g.clock = clock
	var err error
	g.synth, err = synths.NewSynth(playdate.WaveformTriangle)
	if err != nil {
		return err
	}
	// Long attack and decay make the curvature difference intentionally obvious
	// in an audible acceptance scene.
	if err = g.synth.SetEnvelope(1.5, 1, .15, .6); err != nil {
		return err
	}
	if err = g.synth.SetEnvelopeCurvature(g.curvature); err != nil {
		return err
	}
	if err = g.synth.SetEnvelopeVelocitySensitivity(.6); err != nil {
		return err
	}
	if err = g.synth.SetEnvelopeRateScaling(.25, 36, 84); err != nil {
		return err
	}
	return nil
}

func (g *game) Update(context playdate.Context) (bool, error) {
	input := context.Input()
	if input.Pressed.Has(playdate.ButtonA) {
		now, err := g.clock.CurrentAudioTime()
		if err != nil {
			return false, err
		}
		if err = g.synth.PlayMIDINote(60, .9, 3.5, now); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if input.Pressed.Has(playdate.ButtonLeft) || input.Pressed.Has(playdate.ButtonRight) {
		if input.Pressed.Has(playdate.ButtonLeft) {
			g.curvature -= .5
		} else {
			g.curvature += .5
		}
		if g.curvature < -1 {
			g.curvature = -1
		}
		if g.curvature > 1 {
			g.curvature = 1
		}
		if err := g.synth.SetEnvelopeCurvature(g.curvature); err != nil {
			return false, err
		}
		g.dirty = true
	}
	if !g.dirty {
		return false, nil
	}
	g.dirty = false
	context.Clear()
	context.DrawText("P9.3 synthesis", 12, 20)
	context.DrawText("Tap A after each change", 12, 52)
	context.DrawText("Left/Right: curve by 0.5", 12, 76)
	context.DrawText("Curve: "+curveName(g.curvature), 12, 112)
	return true, nil
}

func (g *game) HandleLifecycle(_ playdate.Context, event playdate.LifecycleEvent) error {
	if event == playdate.LifecycleTerminate {
		return g.synth.Close()
	}
	return nil
}

func curveName(v float32) string {
	value := int(v*10 + .5)
	if v < 0 {
		value = int(v*10 - .5)
	}
	if value == 0 {
		return "0.0 (linear)"
	}
	sign := "+"
	if value < 0 {
		sign = "-"
		value = -value
	}
	if value == 10 {
		return sign + "1.0"
	}
	return sign + "0." + string(rune('0'+value))
}
