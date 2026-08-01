package playdate

// Color identifies a solid Playdate bitmap color.
type Color uint8

const (
	ColorClear Color = iota
	ColorWhite
	ColorBlack
)

// Graphics exposes Playdate drawing and bitmap creation services.
type Graphics interface {
	Clear()
	DrawText(text string, x, y int)
	LoadBitmap(path string) (Bitmap, error)
	LoadBitmapTable(path string) (BitmapTable, error)
	NewBitmap(width, height int) (Bitmap, error)
	DrawBitmap(bitmap Bitmap, x, y int) error
	DrawScaledBitmap(bitmap Bitmap, x, y int, scaleX, scaleY float32) error
}

// FontGraphics is the optional custom-font capability implemented by native
// Playdate contexts. Keeping it narrow lets games and tests depend only on the
// font slice when they use it.
type FontGraphics interface {
	LoadFont(path string) (Font, error)
	DrawTextFont(font Font, text string, x, y int) error
}
