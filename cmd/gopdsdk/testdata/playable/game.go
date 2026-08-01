// Package game implements the P1.3 external playable consumer.
package game

import (
	"errors"
	"strconv"

	"github.com/Djunichi/gopdsdk/playdate"
)

const soakSeconds = float32(60)

type phase uint8

const (
	playing phase = iota
	paused
	complete
)

// State is the pure-Go gameplay state.
type State struct {
	PlayerX float32
	TargetX int
	Score   int
	Elapsed float32
	Phase   phase
}

// Frame contains the gameplay-relevant part of one input snapshot.
type Frame struct {
	CrankDelta, DeltaSeconds float32
	Pressed                  playdate.Buttons
}

// Step advances gameplay without runtime or graphics dependencies.
func Step(state State, frame Frame) State {
	if frame.Pressed.Has(playdate.ButtonB) {
		return initialState()
	}
	if state.Phase != playing {
		return state
	}
	state.Elapsed += frame.DeltaSeconds
	state.PlayerX = clamp(state.PlayerX+frame.CrankDelta*0.75, 8, 360)
	if frame.Pressed.Has(playdate.ButtonLeft) {
		state.PlayerX = clamp(state.PlayerX-12, 8, 360)
	}
	if frame.Pressed.Has(playdate.ButtonRight) {
		state.PlayerX = clamp(state.PlayerX+12, 8, 360)
	}
	if frame.Pressed.Has(playdate.ButtonA) && distance(int(state.PlayerX), state.TargetX) <= 24 {
		state.Score++
		state.TargetX = 32 + (state.Score*83)%320
	}
	if state.Elapsed >= soakSeconds {
		state.Phase = complete
	}
	return state
}

func clamp(value, minimum, maximum float32) float32 {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func distance(left, right int) int {
	if left < right {
		return right - left
	}
	return left - right
}

type drawKind uint8

const (
	drawText drawKind = iota
	drawPlayer
	drawTarget
)

// DrawCommand is one deterministic rendering operation.
type DrawCommand struct {
	Kind drawKind
	Text string
	X, Y int
}

// DrawPlan derives the complete frame rendering from state.
func DrawPlan(state State) []DrawCommand {
	status := "PLAY"
	if state.Phase == paused {
		status = "PAUSE"
	}
	if state.Phase == complete {
		status = "PASS"
	}
	return []DrawCommand{
		{Kind: drawText, Text: "P1.3 CRANK CATCH", X: 12, Y: 8},
		{Kind: drawText, Text: "A:catch B:reset d-pad:nudge", X: 12, Y: 30},
		{Kind: drawText, Text: "Score:" + strconv.Itoa(state.Score) + "  " + status + " " + seconds(state.Elapsed) + "/60s", X: 12, Y: 52},
		{Kind: drawTarget, X: state.TargetX, Y: 112},
		{Kind: drawPlayer, X: int(state.PlayerX), Y: 184},
	}
}

func seconds(value float32) string { return strconv.FormatFloat(float64(value), 'f', 1, 32) }

type game struct {
	state          State
	player, target playdate.Bitmap
	closed         bool
}

// New creates the playable consumer entry point expected by gopdsdk.
func New() playdate.Game { return &game{state: initialState()} }

func initialState() State { return State{PlayerX: 184, TargetX: 184} }

func (g *game) Init(context playdate.Context) error {
	player, err := context.LoadBitmap("images/player")
	if err != nil {
		return err
	}
	g.player = player
	target, err := context.LoadBitmap("images/target")
	if err != nil {
		return errors.Join(err, g.close())
	}
	g.target = target
	return nil
}

func (g *game) Update(context playdate.Context) (bool, error) {
	input := context.Input()
	g.state = Step(g.state, Frame{CrankDelta: input.CrankDelta, DeltaSeconds: input.DeltaSeconds, Pressed: input.Pressed})
	context.Clear()
	for _, command := range DrawPlan(g.state) {
		switch command.Kind {
		case drawText:
			context.DrawText(command.Text, command.X, command.Y)
		case drawPlayer:
			if err := context.DrawBitmap(g.player, command.X, command.Y); err != nil {
				return false, err
			}
		case drawTarget:
			if err := context.DrawBitmap(g.target, command.X, command.Y); err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

func (g *game) HandleLifecycle(_ playdate.Context, event playdate.LifecycleEvent) error {
	switch event {
	case playdate.LifecyclePause, playdate.LifecycleLock:
		if g.state.Phase == playing {
			g.state.Phase = paused
		}
	case playdate.LifecycleResume, playdate.LifecycleUnlock:
		if g.state.Phase == paused {
			g.state.Phase = playing
		}
	case playdate.LifecycleTerminate:
		return g.close()
	}
	return nil
}

func (g *game) close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	var err error
	if g.player != nil {
		err = g.player.Close()
	}
	if g.target != nil {
		err = errors.Join(err, g.target.Close())
	}
	return err
}
