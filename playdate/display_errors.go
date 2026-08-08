package playdate

import "errors"

var (
	// ErrDisplayUnavailable indicates a context without display controls.
	ErrDisplayUnavailable = errors.New("display controls are unavailable")
	// ErrDisplayRefreshRate indicates a non-finite rate or a rate outside 0..50 fps.
	ErrDisplayRefreshRate = errors.New("display refresh rate must be finite and between 0 and 50")
	// ErrDisplayScale indicates a scale other than 1, 2, 4, or 8.
	ErrDisplayScale = errors.New("display scale must be 1, 2, 4, or 8")
	// ErrDisplayMosaic indicates a mosaic component outside 0..3.
	ErrDisplayMosaic = errors.New("display mosaic components must be between 0 and 3")
)
