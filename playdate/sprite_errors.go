package playdate

// Sprite errors.

type spriteError string

func (message spriteError) Error() string { return string(message) }

var (
	// ErrSpriteClosed indicates an operation on an already closed sprite.
	ErrSpriteClosed error = spriteError("sprite is closed")
	// ErrSpriteBorrowed indicates an attempt to close a borrowed query result.
	ErrSpriteBorrowed error = spriteError("borrowed sprite cannot be closed")
	// ErrSpriteCreate indicates that Playdate could not allocate a sprite.
	ErrSpriteCreate error = spriteError("create sprite failed")
	// ErrSpriteDisplayListUnavailable indicates a context without P8.2 global controls.
	ErrSpriteDisplayListUnavailable error = spriteError("sprite display-list controls are unavailable")
	// ErrSpriteDirtyRect indicates a dirty rectangle with invalid dimensions or coordinates.
	ErrSpriteDirtyRect error = spriteError("invalid sprite dirty rectangle")
	// ErrSpriteRedrawUnavailable indicates a context without redraw controls.
	ErrSpriteRedrawUnavailable error = spriteError("sprite redraw controls are unavailable")
	// ErrSpriteTileMapCreate indicates that Playdate could not allocate a native tilemap.
	ErrSpriteTileMapCreate error = spriteError("create sprite tilemap failed")
	// ErrSpriteTileMapUnavailable indicates a context without native tilemap creation.
	ErrSpriteTileMapUnavailable error = spriteError("sprite tilemaps are unavailable")
	// ErrSpriteTileMapClosed indicates an operation on a closed native tilemap.
	ErrSpriteTileMapClosed error = spriteError("sprite tilemap is closed")
	// ErrSpriteTileMapBorrowed indicates an attempt to close a borrowed native tilemap.
	ErrSpriteTileMapBorrowed error = spriteError("borrowed sprite tilemap cannot be closed")
	// ErrSpriteTileMapConfig indicates invalid dimensions or tile data.
	ErrSpriteTileMapConfig error = spriteError("invalid sprite tilemap configuration")
	// ErrSpriteTileMapBounds indicates a tile coordinate outside the map.
	ErrSpriteTileMapBounds error = spriteError("sprite tilemap coordinate is out of range")
	// ErrSpriteTileMapInUse indicates a tilemap still attached to a sprite.
	ErrSpriteTileMapInUse error = spriteError("sprite tilemap is still attached")
)
