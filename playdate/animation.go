package playdate

import "math"

// Animation advances a bounded sequence of bitmap-table frames without
// allocating in the update loop.
type Animation struct {
	table        BitmapTable
	first, count int
	frame        int
	frameSeconds float32
	elapsed      float32
	paused       bool
	fixed        bool
}

// NewAnimation creates a looping table-backed animation.
func NewAnimation(table BitmapTable, first, count int, frameSeconds float32) (*Animation, error) {
	if table == nil || first < 0 || count <= 0 || frameSeconds <= 0 || math.IsNaN(float64(frameSeconds)) || math.IsInf(float64(frameSeconds), 0) {
		return nil, ErrAnimationConfig
	}
	return &Animation{table: table, first: first, count: count, frameSeconds: frameSeconds}, nil
}

// Update advances using elapsed seconds. Paused and fixed-frame animations retain state.
func (a *Animation) Update(deltaSeconds float32) {
	if a == nil || a.paused || a.fixed || deltaSeconds <= 0 || math.IsNaN(float64(deltaSeconds)) || math.IsInf(float64(deltaSeconds), 0) {
		return
	}
	a.elapsed += deltaSeconds
	cycleSeconds := a.frameSeconds * float32(a.count)
	if a.elapsed >= cycleSeconds {
		a.elapsed = float32(math.Mod(float64(a.elapsed), float64(cycleSeconds)))
	}
	steps := int(a.elapsed / a.frameSeconds)
	if steps == 0 {
		return
	}
	a.elapsed -= float32(steps) * a.frameSeconds
	a.frame = (a.frame + steps) % a.count
}

// Bitmap returns the current borrowed frame.
func (a *Animation) Bitmap() (Bitmap, error) {
	if a == nil || a.table == nil {
		return nil, ErrAnimationConfig
	}
	return a.table.Frame(a.first + a.frame)
}

// Frame returns the current zero-based frame within the configured sequence.
func (a *Animation) Frame() int {
	if a == nil {
		return 0
	}
	return a.frame
}

// SetFixedFrame selects a frame and disables time-based advancement.
func (a *Animation) SetFixedFrame(frame int) error {
	if a == nil || frame < 0 || frame >= a.count {
		return ErrBitmapFrameRange
	}
	a.frame, a.fixed, a.elapsed = frame, true, 0
	return nil
}

// UseDeltaTime leaves fixed-frame mode without resetting the selected frame.
func (a *Animation) UseDeltaTime() {
	if a != nil {
		a.fixed = false
	}
}

// Pause stops advancement without discarding partial-frame time.
func (a *Animation) Pause() {
	if a != nil {
		a.paused = true
	}
}

// Resume continues advancement from the retained state.
func (a *Animation) Resume() {
	if a != nil {
		a.paused = false
	}
}

// Paused reports whether advancement is paused.
func (a *Animation) Paused() bool { return a != nil && a.paused }
