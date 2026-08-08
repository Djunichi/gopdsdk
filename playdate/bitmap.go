package playdate

// Bitmap is a Playdate bitmap handle. Only owned handles may be closed.
type Bitmap interface {
	Width() (int, error)
	Height() (int, error)
	Clear() error
	Fill(Color) error
	Close() error
}

// BitmapTable owns a Playdate bitmap table. Frames borrowed from it remain
// valid only until the table is closed.
type BitmapTable interface {
	Frame(index int) (Bitmap, error)
	Close() error
}

// BitmapFlip selects the reflection applied while testing bitmap masks.
type BitmapFlip uint8

const (
	BitmapUnflipped BitmapFlip = iota
	BitmapFlippedX
	BitmapFlippedY
	BitmapFlippedXY
)

// BitmapData is a callback-scoped view of a bitmap's one-bit image and mask.
// The byte slices alias native memory and expire when the callback returns.
type BitmapData interface {
	Width() int
	Height() int
	RowBytes() int
	Bytes() ([]byte, error)
	MaskBytes() ([]byte, error)
	Dirty() (bool, error)
	MarkDirty() error
	Pixel(x, y int) (Color, error)
	SetPixel(x, y int, color Color) error
}

// BitmapDataGraphics contains the owned-bitmap, mask, collision, and snapshot
// operations that are separate from immediate-mode drawing.
type BitmapDataGraphics interface {
	WithBitmapData(bitmap Bitmap, callback func(BitmapData) error) error
	CopyBitmap(bitmap Bitmap) (Bitmap, error)
	LoadIntoBitmap(path string, bitmap Bitmap) error
	NewBitmapTable(count, width, height int) (BitmapTable, error)
	LoadIntoBitmapTable(path string, table BitmapTable) error
	SetBitmapMask(bitmap, mask Bitmap) error
	ClearBitmapMask(bitmap Bitmap) error
	// BitmapMask returns an owned mask view that must be closed before bitmap.
	BitmapMask(bitmap Bitmap) (Bitmap, bool, error)
	CheckBitmapMaskCollision(first Bitmap, firstX, firstY int, firstFlip BitmapFlip, second Bitmap, secondX, secondY int, secondFlip BitmapFlip, rectX, rectY, rectWidth, rectHeight int) (bool, error)
	RotatedBitmap(bitmap Bitmap, degrees, scaleX, scaleY float32) (Bitmap, int, error)
	CopyDisplayBuffer() (Bitmap, error)
}
