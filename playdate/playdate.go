// Package playdate defines the public application contract exposed by gopdsdk.
package playdate

// Game receives lifecycle callbacks from the Playdate runtime.
type Game interface {
	Init(Context) error
	Update(Context) (refresh bool, err error)
}

// LifecycleGame optionally receives system lifecycle events. Events are
// delivered in platform order and an error permanently stops the application.
type LifecycleGame interface {
	HandleLifecycle(Context, LifecycleEvent) error
}

// LifecycleEvent identifies a lifecycle transition visible to game code.
type LifecycleEvent uint8

const (
	LifecyclePause LifecycleEvent = iota
	LifecycleResume
	LifecycleLock
	LifecycleUnlock
	LifecycleTerminate
	LifecycleLowPower
)

// Buttons is a bit set of Playdate controls.
type Buttons uint8

const (
	ButtonLeft Buttons = 1 << iota
	ButtonRight
	ButtonUp
	ButtonDown
	ButtonB
	ButtonA
)

// Has reports whether every requested button is present in the set.
func (buttons Buttons) Has(requested Buttons) bool { return buttons&requested == requested }

// Input is the immutable input snapshot for the current frame.
type Input struct {
	Buttons              Buttons
	Pressed              Buttons
	Released             Buttons
	Held                 Buttons
	CrankAngle           float32
	CrankDelta           float32
	CrankDocked          bool
	CrankDockedThisFrame bool
	CrankUndocked        bool
	DeltaSeconds         float32
}

// Context exposes Playdate services available to a game callback.
type Context interface {
	Clear()
	DrawText(text string, x, y int)
	// CurrentTimeMilliseconds returns the wrapping monotonic device clock.
	CurrentTimeMilliseconds() uint32
	// Input returns the snapshot captured at the start of the current frame.
	Input() Input
}
