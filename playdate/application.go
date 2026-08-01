// Package playdate defines the public application contract exposed by gopdsdk.
package playdate

// Game receives lifecycle callbacks from the Playdate runtime.
type Game interface {
	Init(Context) error
	Update(Context) (refresh bool, err error)
}

// System exposes Playdate system services.
type System interface {
	// CurrentTimeMilliseconds returns the wrapping monotonic device clock.
	CurrentTimeMilliseconds() uint32
}

// Context exposes Playdate capabilities available to a game callback.
type Context interface {
	System
	Graphics
	InputReader
	Sprites
	Audio
}
