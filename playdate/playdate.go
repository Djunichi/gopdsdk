// Package playdate defines the public application contract exposed by gopdsdk.
package playdate

// Game receives lifecycle callbacks from the Playdate runtime.
type Game interface {
	Init(Context) error
	Update(Context) (refresh bool, err error)
}

// Context exposes Playdate services available to a game callback.
type Context interface {
	DrawText(text string, x, y int)
}
