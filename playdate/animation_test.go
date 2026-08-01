package playdate

import (
	"errors"
	"testing"
)

type animationTable struct{ frames []Bitmap }

func (t animationTable) Frame(index int) (Bitmap, error) {
	if index < 0 || index >= len(t.frames) {
		return nil, ErrBitmapFrameRange
	}
	return t.frames[index], nil
}
func (animationTable) Close() error { return nil }

func TestAnimationDeltaPauseAndFixedFrame(t *testing.T) {
	table := animationTable{frames: make([]Bitmap, 4)}
	a, err := NewAnimation(table, 0, 4, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	a.Update(0.25)
	if a.Frame() != 2 {
		t.Fatalf("frame = %d, want 2", a.Frame())
	}
	a.Pause()
	a.Update(1)
	a.Resume()
	a.Update(0.04)
	if a.Frame() != 2 {
		t.Fatalf("paused state lost: frame = %d", a.Frame())
	}
	a.Update(0.02)
	if a.Frame() != 3 {
		t.Fatalf("retained delta lost: frame = %d", a.Frame())
	}
	if err := a.SetFixedFrame(1); err != nil {
		t.Fatal(err)
	}
	a.Update(10)
	if a.Frame() != 1 {
		t.Fatalf("fixed frame advanced to %d", a.Frame())
	}
	a.UseDeltaTime()
	a.Update(0.1)
	if a.Frame() != 2 {
		t.Fatalf("delta mode frame = %d", a.Frame())
	}
	if err := a.SetFixedFrame(4); !errors.Is(err, ErrBitmapFrameRange) {
		t.Fatalf("range error = %v", err)
	}
}

func TestAnimationRejectsInvalidConfig(t *testing.T) {
	if _, err := NewAnimation(nil, 0, 1, 1); !errors.Is(err, ErrAnimationConfig) {
		t.Fatalf("error = %v", err)
	}
}

func TestAnimationUpdateDoesNotAllocate(t *testing.T) {
	a, err := NewAnimation(animationTable{frames: make([]Bitmap, 2)}, 0, 2, 0.1)
	if err != nil {
		t.Fatal(err)
	}
	if allocations := testing.AllocsPerRun(100, func() { a.Update(0.016) }); allocations != 0 {
		t.Fatalf("allocations per update = %v, want 0", allocations)
	}
}
