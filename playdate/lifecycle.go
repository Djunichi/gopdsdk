package playdate

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
