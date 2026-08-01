package game

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Djunichi/gopdsdk/playdate"
)

func TestStepUsesCrankButtonsDeltaAndCatch(t *testing.T) {
	got := Step(State{PlayerX: 100, TargetX: 124}, Frame{CrankDelta: 16, DeltaSeconds: 0.25, Pressed: playdate.ButtonRight | playdate.ButtonA})
	want := State{PlayerX: 124, TargetX: 115, Score: 1, Elapsed: 0.25}
	if got != want {
		t.Fatalf("Step() = %+v, want %+v", got, want)
	}
}

func TestStepClampsAndCompletesSixtySecondSoak(t *testing.T) {
	got := Step(State{PlayerX: 359, TargetX: 10, Elapsed: 59.5}, Frame{CrankDelta: 100, DeltaSeconds: 0.5})
	if got.PlayerX != 360 || got.Phase != complete || got.Elapsed != 60 {
		t.Fatalf("Step() = %+v", got)
	}
	if after := Step(got, Frame{CrankDelta: -100, DeltaSeconds: 1}); after != got {
		t.Fatalf("complete state changed: %+v", after)
	}
}

func TestDrawPlanIsDeterministic(t *testing.T) {
	state := State{PlayerX: 42.9, TargetX: 123, Score: 2, Elapsed: 60, Phase: complete}
	want := []DrawCommand{{drawText, "P1.3 CRANK CATCH", 12, 8}, {drawText, "A:catch B:reset d-pad:nudge", 12, 30}, {drawText, "Score:2  PASS 60.0/60s", 12, 52}, {drawTarget, "", 123, 112}, {drawPlayer, "", 42, 184}}
	if got := DrawPlan(state); !reflect.DeepEqual(got, want) {
		t.Fatalf("DrawPlan() = %#v, want %#v", got, want)
	}
}

func TestButtonBResetsState(t *testing.T) {
	changed := State{PlayerX: 42, TargetX: 99, Score: 7, Elapsed: 12}
	if got := Step(changed, Frame{Pressed: playdate.ButtonB}); got != initialState() {
		t.Fatalf("reset = %+v", got)
	}
}

type bitmap struct {
	name       string
	operations *[]string
	closeErr   error
}

func (*bitmap) Width() (int, error)       { return 16, nil }
func (*bitmap) Height() (int, error)      { return 16, nil }
func (*bitmap) Clear() error              { return nil }
func (*bitmap) Fill(playdate.Color) error { return nil }
func (b *bitmap) Close() error {
	*b.operations = append(*b.operations, "close:"+b.name)
	return b.closeErr
}

type context struct {
	input       playdate.Input
	operations  []string
	loadErrorAt int
	loads       int
}

func (*context) CurrentTimeMilliseconds() uint32  { return 0 }
func (c *context) Input() playdate.Input          { return c.input }
func (c *context) Clear()                         { c.operations = append(c.operations, "clear") }
func (c *context) DrawText(text string, x, y int) { c.operations = append(c.operations, "text:"+text) }
func (c *context) LoadBitmap(path string) (playdate.Bitmap, error) {
	c.loads++
	c.operations = append(c.operations, "load:"+path)
	if c.loads == c.loadErrorAt {
		return nil, errors.New("load failed")
	}
	return &bitmap{name: path, operations: &c.operations}, nil
}
func (*context) NewBitmap(int, int) (playdate.Bitmap, error) { return nil, nil }
func (c *context) DrawBitmap(value playdate.Bitmap, x, y int) error {
	c.operations = append(c.operations, "draw:"+value.(*bitmap).name)
	return nil
}
func (*context) DrawScaledBitmap(playdate.Bitmap, int, int, float32, float32) error { return nil }

func TestGameExecutesPlanAndClosesOnce(t *testing.T) {
	c := &context{input: playdate.Input{CrankDelta: 8, DeltaSeconds: 1}}
	g := New().(*game)
	if err := g.Init(c); err != nil {
		t.Fatal(err)
	}
	if refresh, err := g.Update(c); err != nil || !refresh {
		t.Fatalf("Update() = %v, %v", refresh, err)
	}
	if err := g.HandleLifecycle(c, playdate.LifecycleTerminate); err != nil {
		t.Fatal(err)
	}
	if err := g.HandleLifecycle(c, playdate.LifecycleTerminate); err != nil {
		t.Fatal(err)
	}
	want := []string{"load:images/player", "load:images/target", "clear", "text:P1.3 CRANK CATCH", "text:A:catch B:reset d-pad:nudge", "text:Score:0  PLAY 1.0/60s", "draw:images/target", "draw:images/player", "close:images/player", "close:images/target"}
	if !reflect.DeepEqual(c.operations, want) {
		t.Fatalf("operations = %v, want %v", c.operations, want)
	}
}

func TestInitRollsBackFirstBitmap(t *testing.T) {
	c := &context{loadErrorAt: 2}
	if err := New().Init(c); err == nil {
		t.Fatal("Init() succeeded")
	}
	want := []string{"load:images/player", "load:images/target", "close:images/player"}
	if !reflect.DeepEqual(c.operations, want) {
		t.Fatalf("operations = %v, want %v", c.operations, want)
	}
}

func TestLifecyclePausesGameplay(t *testing.T) {
	g := New().(*game)
	c := &context{}
	if err := g.HandleLifecycle(c, playdate.LifecyclePause); err != nil {
		t.Fatal(err)
	}
	before := g.state
	g.state = Step(g.state, Frame{CrankDelta: 20, DeltaSeconds: 1})
	if g.state != before {
		t.Fatalf("paused state changed: %+v", g.state)
	}
	if err := g.HandleLifecycle(c, playdate.LifecycleResume); err != nil {
		t.Fatal(err)
	}
	if g.state.Phase != playing {
		t.Fatalf("phase = %v", g.state.Phase)
	}
}
