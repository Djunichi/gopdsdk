package playdate

// Sprite is an explicitly owned Playdate display-list object. Closing a sprite
// removes it from the display list before releasing its native handle.
type Sprite interface {
	SetBitmap(Bitmap) error
	SetPosition(x, y float32) error
	MoveBy(dx, dy float32) error
	SetVisible(bool) error
	SetZIndex(int) error
	Add() error
	Remove() error
	Close() error
}

// Sprites exposes sprite creation and the global display-list frame pass.
type Sprites interface {
	NewSprite() (Sprite, error)
	UpdateAndDrawSprites()
}
