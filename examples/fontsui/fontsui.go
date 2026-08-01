// Package fontsui is the P2.5 custom-font and deterministic game-UI acceptance scene.
package fontsui

import (
	"errors"
	"strconv"

	"github.com/Djunichi/gopdsdk/playdate"
)

const screenWidth = 400

// Phase identifies the visible game screen.
type Phase uint8

const (
	Playing Phase = iota
	Paused
	GameOver
)

// State is all input consumed by the deterministic UI plan.
type State struct {
	Phase Phase
	Score int
}

// TextMeasurer is the measurement dependency used by LayoutPlan.
type TextMeasurer interface {
	TextWidth(string) (int, error)
	Height() (int, error)
}

// TextCommand is one fully positioned custom-font draw operation.
type TextCommand struct {
	Text string
	X, Y int
}

// LayoutPlan derives a deterministic HUD or overlay from state and font metrics.
func LayoutPlan(state State, font TextMeasurer) ([]TextCommand, error) {
	height, err := font.Height()
	if err != nil {
		return nil, err
	}
	score := "SCORE " + strconv.Itoa(state.Score)
	result := []TextCommand{{Text: "P2.5 FONTS + UI", X: 12, Y: 10}, {Text: score, X: 12, Y: 10 + height + 4}}

	var title, hint string
	switch state.Phase {
	case Playing:
		title, hint = "HUD", "A:SCORE  B:GAME OVER"
	case Paused:
		title, hint = "PAUSED", "MENU:RESUME"
	case GameOver:
		title, hint = "GAME OVER", "A:RESTART"
	default:
		return nil, errors.New("unknown UI phase")
	}
	for index, line := range []string{title, hint} {
		width, measureErr := font.TextWidth(line)
		if measureErr != nil {
			return nil, measureErr
		}
		result = append(result, TextCommand{Text: line, X: (screenWidth - width) / 2, Y: 104 + index*(height+8)})
	}
	return result, nil
}

type game struct {
	state     State
	font      playdate.Font
	fontError error
}

// New creates the P2.5 acceptance game.
func New() playdate.Game { return &game{} }

func (g *game) Init(context playdate.Context) error {
	graphics, ok := context.(playdate.FontGraphics)
	if !ok {
		return errors.New("custom-font graphics are unavailable")
	}
	font, err := graphics.LoadFont("fonts/gopdsdk-ui")
	if err != nil {
		g.fontError = err
		g.state = State{Phase: Playing}
		return nil
	}
	g.font = font
	g.state = State{Phase: Playing}
	return nil
}

func (g *game) Update(context playdate.Context) (bool, error) {
	context.Clear()
	if g.fontError != nil {
		context.DrawText("FONT ERROR: "+g.fontError.Error(), 12, 12)
		return true, nil
	}
	input := context.Input()
	if g.state.Phase == GameOver && input.Pressed.Has(playdate.ButtonA) {
		g.state = State{Phase: Playing}
	} else if g.state.Phase == Playing {
		if input.Pressed.Has(playdate.ButtonA) {
			g.state.Score++
		}
		if input.Pressed.Has(playdate.ButtonB) {
			g.state.Phase = GameOver
		}
	}
	plan, err := LayoutPlan(g.state, g.font)
	if err != nil {
		return false, err
	}
	graphics := context.(playdate.FontGraphics)
	for _, command := range plan {
		if err := graphics.DrawTextFont(g.font, command.Text, command.X, command.Y); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (g *game) HandleLifecycle(_ playdate.Context, event playdate.LifecycleEvent) error {
	switch event {
	case playdate.LifecyclePause, playdate.LifecycleLock:
		if g.state.Phase == Playing {
			g.state.Phase = Paused
		}
	case playdate.LifecycleResume, playdate.LifecycleUnlock:
		if g.state.Phase == Paused {
			g.state.Phase = Playing
		}
	case playdate.LifecycleTerminate:
		if g.font != nil {
			err := g.font.Close()
			g.font = nil
			return err
		}
	}
	return nil
}
