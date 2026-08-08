package playdate

// Display is the optional physical-display presentation capability.
type Display interface {
	Width() int
	Height() int
	RefreshRate() float32
	FPS() float32
	SetRefreshRate(framesPerSecond float32) error
	SetInverted(bool)
	SetScale(uint) error
	SetMosaic(x, y uint) error
	SetFlipped(x, y bool)
	SetOffset(x, y int)
}
