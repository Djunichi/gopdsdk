package playdate

// Sprite is an explicitly owned Playdate display-list object. Closing a sprite
// removes it from the display list before releasing its native handle.
type Sprite interface {
	SetBitmap(Bitmap) error
	SetPosition(x, y float32) error
	MoveBy(dx, dy float32) error
	SetVisible(bool) error
	SetZIndex(int) error
	SetCollideRect(Rect) error
	ClearCollideRect() error
	SetTag(uint8) error
	MarkDirty() error
	MarkDirtyRect(Rect) error
	MoveWithCollisions(goalX, goalY float32) (MoveResult, error)
	Add() error
	Remove() error
	Close() error
}

// Sprites exposes sprite creation and the global display-list frame pass.
type Sprites interface {
	NewSprite() (Sprite, error)
	QuerySpritesAtPoint(x, y float32) []Sprite
	QuerySpritesInRect(Rect) []Sprite
	QueryOverlappingSprites(Sprite) ([]Sprite, error)
	UpdateAndDrawSprites()
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
