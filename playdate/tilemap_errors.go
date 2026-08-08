package playdate

// Tile-map errors.

type tileMapError string

func (message tileMapError) Error() string { return string(message) }

var (
	// ErrTileMapConfig indicates invalid layer dimensions or tile data length.
	ErrTileMapConfig error = tileMapError("invalid tile map configuration")
	// ErrTileMapBitmap indicates that a visible tile has no corresponding bitmap.
	ErrTileMapBitmap error = tileMapError("tile bitmap is unavailable")
	// ErrTileMapDraw indicates invalid drawing inputs or an underlying draw failure.
	ErrTileMapDraw error = tileMapError("tile map draw failed")
)
