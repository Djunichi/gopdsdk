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
