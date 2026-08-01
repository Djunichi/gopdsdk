package playdate

// Bitmap is a Playdate bitmap handle. Only owned handles may be closed.
type Bitmap interface {
	Width() (int, error)
	Height() (int, error)
	Clear() error
	Fill(Color) error
	Close() error
}
