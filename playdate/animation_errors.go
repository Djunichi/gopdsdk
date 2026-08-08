package playdate

// Animation errors.

type animationError string

func (message animationError) Error() string { return string(message) }

// ErrAnimationConfig indicates invalid animation timing or frame bounds.
var ErrAnimationConfig error = animationError("invalid animation configuration")
