package playdate

// Graphics errors.

type graphicsError string

func (message graphicsError) Error() string { return string(message) }

var (
	// ErrGraphicsColor indicates a solid color outside the public values.
	ErrGraphicsColor error = graphicsError("invalid graphics color")
	// ErrGraphicsUnavailable indicates a context without the optional graphics slice.
	ErrGraphicsUnavailable error = graphicsError("graphics capability is unavailable")
	// ErrGraphicsGeometry indicates non-positive dimensions, widths, or non-finite angles.
	ErrGraphicsGeometry error = graphicsError("invalid graphics geometry")
	// ErrGraphicsDrawMode indicates a draw mode outside the public values.
	ErrGraphicsDrawMode error = graphicsError("invalid graphics draw mode")
	// ErrGraphicsLineCap indicates a line-cap style outside the public values.
	ErrGraphicsLineCap error = graphicsError("invalid graphics line cap style")
	// ErrGraphicsPolygon indicates fewer than three vertices or an invalid fill rule.
	ErrGraphicsPolygon error = graphicsError("invalid graphics polygon")
	// ErrGraphicsStencilCallback indicates a nil scoped stencil callback.
	ErrGraphicsStencilCallback error = graphicsError("stencil callback is required")
	// ErrGraphicsStencilActive indicates a nested scoped stencil operation.
	ErrGraphicsStencilActive error = graphicsError("stencil callback is already active")
	// ErrGraphicsStencilWidth indicates a tiled stencil whose width is not a multiple of 32.
	ErrGraphicsStencilWidth error = graphicsError("tiled stencil width must be a multiple of 32")
)
