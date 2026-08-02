package playdate

// Color identifies a solid Playdate bitmap color.
type Color uint8

const (
	ColorClear Color = iota
	ColorWhite
	ColorBlack
)

// DrawMode controls how source bitmap pixels combine with the destination.
type DrawMode uint8

const (
	DrawModeCopy DrawMode = iota
	DrawModeWhiteTransparent
	DrawModeBlackTransparent
	DrawModeFillWhite
	DrawModeFillBlack
	DrawModeXOR
	DrawModeNXOR
	DrawModeInverted
)

// Paint is either a solid drawing color, XOR, or an 8x8 image-and-mask pattern.
type Paint struct {
	pattern [16]byte
	solid   Color
	kind    uint8
}

// SolidPaint creates paint for a public solid color.
func SolidPaint(color Color) (Paint, error) {
	if color > ColorBlack {
		return Paint{}, ErrGraphicsColor
	}
	return Paint{solid: color}, nil
}

// XORPaint creates paint that toggles destination pixels.
func XORPaint() Paint { return Paint{kind: 1} }

// PatternPaint copies an 8x8 image and mask into a self-contained paint value.
func PatternPaint(image, mask [8]byte) Paint {
	paint := Paint{kind: 2}
	copy(paint.pattern[:8], image[:])
	copy(paint.pattern[8:], mask[:])
	return paint
}

// Components returns the portable adapter representation of paint.
func (paint Paint) Components() (solid uint8, pattern [16]byte, patterned bool) {
	switch paint.kind {
	case 1:
		return 3, pattern, false
	case 2:
		return 0, paint.pattern, true
	default:
		return uint8(paint.solid), pattern, false
	}
}

// PrimitiveGraphics is the optional immediate-mode primitive drawing slice.
type PrimitiveGraphics interface {
	DrawLine(x1, y1, x2, y2, width int, paint Paint) error
	DrawRect(x, y, width, height int, paint Paint) error
	FillRect(x, y, width, height int, paint Paint) error
	DrawEllipse(x, y, width, height, lineWidth int, startAngle, endAngle float32, paint Paint) error
	FillEllipse(x, y, width, height int, startAngle, endAngle float32, paint Paint) error
	DrawTriangle(x1, y1, x2, y2, x3, y3, width int, paint Paint) error
	FillTriangle(x1, y1, x2, y2, x3, y3 int, paint Paint) error
}

// GraphicsState is the optional clipping, offset, and draw-mode slice.
type GraphicsState interface {
	SetClipRect(x, y, width, height int) error
	ClearClipRect()
	SetDrawOffset(dx, dy int)
	SetDrawMode(mode DrawMode) error
}

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
