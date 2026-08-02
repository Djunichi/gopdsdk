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

// Launcher is the optional system capability for stopping the current game and
// returning to the Playdate Launcher. The runtime delivers LifecycleTerminate
// before the Launcher starts so games can release their owned resources.
type Launcher interface {
	ExitToLauncher()
}

// Context exposes Playdate capabilities available to a game callback.
type Context interface {
	System
	Graphics
	InputReader
	Sprites
	Audio
}
