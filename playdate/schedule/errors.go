package schedule

type scheduleError string

func (message scheduleError) Error() string { return string(message) }

var (
	// ErrConfig indicates invalid scheduler or queue bounds.
	ErrConfig error = scheduleError("invalid schedule configuration")
	// ErrTask indicates a nil task step.
	ErrTask error = scheduleError("task step is nil")
	// ErrFull indicates that fixed scheduler capacity is exhausted.
	ErrFull error = scheduleError("scheduler capacity is full")
	// ErrDelay indicates an ambiguous wrapping-clock interval or zero ticker interval.
	ErrDelay error = scheduleError("task delay is outside the supported range")
	// ErrTerminated indicates use after scheduler termination.
	ErrTerminated error = scheduleError("scheduler is terminated")
	// ErrUpdating indicates a nested Update call from a task step.
	ErrUpdating error = scheduleError("scheduler update is already running")
)
