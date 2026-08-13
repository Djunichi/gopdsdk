package playdate

// ButtonCallbackQueueLimit is the largest native button-event queue accepted
// by SetButtonCallback. Events are delivered in platform order immediately
// before the next game update.
const ButtonCallbackQueueLimit = 64

// ButtonEvent is one native button transition retained between update cycles.
// When uses the wrapping monotonic device clock in milliseconds.
type ButtonEvent struct {
	Button Buttons
	Down   bool
	When   uint32
}

// ButtonCallback receives every retained native button transition. Unlike the
// frame Input snapshot, it preserves multiple transitions of the same button
// between updates.
type ButtonCallback func(ButtonEvent)

// SystemControls is the optional launch and lifecycle-control capability.
// Runtime-owned settings, callbacks, and menu-image references are restored or
// cleared before LifecycleTerminate is delivered to the game.
type SystemControls interface {
	// LaunchArguments returns copied launch arguments and the loaded game path.
	LaunchArguments() (arguments, path string)
	// RestartGame restarts the current game with arguments available to the new
	// instance through LaunchArguments.
	RestartGame(arguments string) error
	// SetMenuImage retains an owned 400x240 bitmap until it is replaced,
	// cleared, or the application terminates. xOffset must be in [0, 200].
	SetMenuImage(bitmap Bitmap, xOffset int) error
	ClearMenuImage()
	// SetAutoLockDisabled changes the three-minute device auto-lock behavior.
	SetAutoLockDisabled(disabled bool)
	// SetCrankSoundsDisabled changes the system crank dock/undock sounds and
	// reports their previous disabled state.
	SetCrankSoundsDisabled(disabled bool) (previous bool)
	// SetButtonCallback replaces the button-event callback. A nil callback with
	// queueSize zero disables delivery. Other queue sizes must be in
	// [1, ButtonCallbackQueueLimit].
	SetButtonCallback(callback ButtonCallback, queueSize int) error
	// ButtonCallbackOverflow returns the number of bridge events dropped since
	// the current callback was installed. The bridge drops newest events when
	// its fixed-capacity queue is full.
	ButtonCallbackOverflow() uint32
}
