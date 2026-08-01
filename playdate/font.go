package playdate

// Font is an owned Playdate font loaded from a packaged resource. Close must
// be called exactly once; drawing and measurement fail after it is closed.
type Font interface {
	TextWidth(text string) (int, error)
	Height() (int, error)
	Close() error
}
