package playdate

type systemError string

func (message systemError) Error() string { return string(message) }

var (
	// ErrDateTime indicates a calendar value outside the Playdate epoch range or
	// with an invalid date or time field.
	ErrDateTime error = systemError("date and time are outside the Playdate epoch range")
	// ErrSystemEnvironmentUnavailable indicates that offline clock, calendar,
	// and device information services are absent.
	ErrSystemEnvironmentUnavailable error = systemError("system environment is unavailable")
	// ErrSystemControlsUnavailable indicates that launch and lifecycle controls
	// are absent or the application is terminating.
	ErrSystemControlsUnavailable error = systemError("system controls are unavailable")
	// ErrLaunchArguments indicates an embedded NUL that cannot cross the native
	// C string boundary.
	ErrLaunchArguments error = systemError("launch arguments contain a NUL byte")
	// ErrMenuImageSize indicates a menu image other than 400x240 pixels.
	ErrMenuImageSize error = systemError("menu image must be 400x240 pixels")
	// ErrMenuImageOffset indicates an x offset outside the official 0..200 range.
	ErrMenuImageOffset error = systemError("menu image x offset must be between 0 and 200")
	// ErrButtonCallbackConfig indicates a missing callback or queue outside the
	// supported bounded range.
	ErrButtonCallbackConfig error = systemError("button callback requires a queue size between 1 and 64")
)
