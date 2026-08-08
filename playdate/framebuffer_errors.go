package playdate

// Framebuffer and offscreen drawing errors.

type framebufferError string

func (message framebufferError) Error() string { return string(message) }

type offscreenError string

func (message offscreenError) Error() string { return string(message) }

var (
	// ErrFramebufferExpired indicates use after the scoped callback returned.
	ErrFramebufferExpired error = framebufferError("framebuffer view has expired")
	// ErrFramebufferBounds indicates a pixel or dirty-row range outside the display.
	ErrFramebufferBounds error = framebufferError("framebuffer coordinates are out of range")
	// ErrFramebufferColor indicates a framebuffer pixel color other than black or white.
	ErrFramebufferColor error = framebufferError("framebuffer pixels must be black or white")
	// ErrFramebufferCallback indicates a nil scoped framebuffer callback.
	ErrFramebufferCallback error = framebufferError("framebuffer callback is required")
	// ErrOffscreenCallback indicates a nil offscreen drawing callback.
	ErrOffscreenCallback error = offscreenError("offscreen callback is required")
)
