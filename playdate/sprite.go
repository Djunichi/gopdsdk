package playdate

// Sprite is an explicitly owned Playdate display-list object. Closing a sprite
// removes it from the display list before releasing its native handle.
type Sprite interface {
	SetBitmap(Bitmap) error
	SetCenter(x, y float32) error
	Center() (x, y float32, err error)
	SetBounds(Rect) error
	Bounds() (Rect, error)
	SetPosition(x, y float32) error
	Position() (x, y float32, err error)
	MoveBy(dx, dy float32) error
	SetVisible(bool) error
	Visible() (bool, error)
	SetZIndex(int) error
	ZIndex() (int, error)
	SetImageFlip(BitmapFlip) error
	ImageFlip() (BitmapFlip, error)
	SetDrawMode(DrawMode) error
	SetOpaque(bool) error
	SetStencilImage(Bitmap, bool) error
	SetStencilPattern([8]byte) error
	ClearStencil() error
	SetClipRect(x, y, width, height int) error
	ClearClipRect() error
	SetIgnoresDrawOffset(bool) error
	SetTileMap(SpriteTileMap) error
	ClearTileMap() error
	TileMap() (SpriteTileMap, bool, error)
	SetUpdatesEnabled(bool) error
	UpdatesEnabled() (bool, error)
	SetCollisionsEnabled(bool) error
	CollisionsEnabled() (bool, error)
	SetCollideRect(Rect) error
	CollideRect() (Rect, error)
	ClearCollideRect() error
	SetTag(uint8) error
	Tag() (uint8, error)
	MarkDirty() error
	MarkDirtyRect(Rect) error
	MoveWithCollisions(goalX, goalY float32) (MoveResult, error)
	CheckCollisions(goalX, goalY float32) (MoveResult, error)
	SetDrawCallback(SpriteDrawCallback) error
	SetUpdateCallback(SpriteUpdateCallback) error
	SetCollisionResponseCallback(SpriteCollisionResponseCallback) error
	Add() error
	Remove() error
	Close() error
}

// SpriteDrawCallback draws a procedural sprite. Bounds is the sprite's current
// bounds and drawRect is the dirty portion requested for this frame. A nil
// callback restores the SDK's default bitmap drawing behavior.
type SpriteDrawCallback func(sprite Sprite, bounds, drawRect Rect)

// SpriteUpdateCallback runs in display-list order during the sprite frame pass.
type SpriteUpdateCallback func(sprite Sprite)

// SpriteCollisionResponseCallback selects the response for one ordered pair.
type SpriteCollisionResponseCallback func(sprite, other Sprite) CollisionResponse

// Sprites exposes sprite creation and the global display-list frame pass.
type Sprites interface {
	NewSprite() (Sprite, error)
	QuerySpritesAtPoint(x, y float32) []Sprite
	QuerySpritesInRect(Rect) []Sprite
	QueryOverlappingSprites(Sprite) ([]Sprite, error)
	UpdateAndDrawSprites()
}

// SpriteQueries exposes line-segment queries and detailed ordered hits.
type SpriteQueries interface {
	QuerySpritesAlongLine(x1, y1, x2, y2 float32) []Sprite
	QuerySpriteInfoAlongLine(x1, y1, x2, y2 float32) []SpriteQueryInfo
}

// SpriteDisplayList exposes global and batch display-list controls.
type SpriteDisplayList interface {
	SpriteCount() int
	AddSprites([]Sprite) error
	RemoveSprites([]Sprite) error
	RemoveAllSprites()
	ResetCollisionWorld()
}

// SpriteTileMaps exposes native tilemap creation when the active runtime
// supports sprite tilemap attachment.
type SpriteTileMaps interface {
	NewSpriteTileMap(BitmapTable, int, int, []uint16) (SpriteTileMap, error)
}

// SpriteTileMap is an explicitly owned native tilemap suitable for attachment
// to a sprite. Close it only after detaching it from every sprite.
type SpriteTileMap interface {
	Size() (columns, rows int, err error)
	PixelSize() (width, height int, err error)
	Tile(column, row int) (uint16, error)
	SetTile(column, row int, index uint16) error
	Close() error
}

// SpriteRedraw is the optional global dirty-region policy capability.
type SpriteRedraw interface {
	SetAlwaysRedraw(bool)
	AddDirtyRect(x, y, width, height int) error
}

// Point is a position or direction in Playdate screen coordinates.
type Point struct{ X, Y float32 }

// Rect is an axis-aligned rectangle in Playdate screen coordinates.
type Rect struct{ X, Y, Width, Height float32 }

// CollisionResponse selects how moveWithCollisions resolves contact.
type CollisionResponse uint8

const (
	CollisionSlide CollisionResponse = iota
	CollisionFreeze
	CollisionOverlap
	CollisionBounce
)

// Collision is the compact portable part of SpriteCollisionInfo.
type Collision struct {
	Other                 Sprite
	ResponseType          CollisionResponse
	Overlaps              bool
	Time                  float32
	Move, Normal, Touch   Point
	SpriteRect, OtherRect Rect
}

// MoveResult reports the resolved position and ordered contacts.
type MoveResult struct {
	ActualX, ActualY float32
	Collisions       []Collision
}

// SpriteQueryInfo describes one ordered intersection with a line segment.
type SpriteQueryInfo struct {
	Sprite                Sprite
	EntryTime, ExitTime   float32
	EntryPoint, ExitPoint Point
}
