package playdate

import "testing"

func TestSentinelErrorDomains(t *testing.T) {
	tests := []struct {
		name string
		err  error
		ok   func(error) bool
	}{
		{"bitmap", ErrBitmapClosed, func(err error) bool { _, ok := err.(bitmapError); return ok }},
		{"graphics", ErrGraphicsGeometry, func(err error) bool { _, ok := err.(graphicsError); return ok }},
		{"framebuffer", ErrFramebufferBounds, func(err error) bool { _, ok := err.(framebufferError); return ok }},
		{"offscreen", ErrOffscreenCallback, func(err error) bool { _, ok := err.(offscreenError); return ok }},
		{"tilemap", ErrTileMapConfig, func(err error) bool { _, ok := err.(tileMapError); return ok }},
		{"animation", ErrAnimationConfig, func(err error) bool { _, ok := err.(animationError); return ok }},
		{"sprite", ErrSpriteClosed, func(err error) bool { _, ok := err.(spriteError); return ok }},
		{"audio", ErrAudioClosed, func(err error) bool { _, ok := err.(audioError); return ok }},
		{"font", ErrFontClosed, func(err error) bool { _, ok := err.(fontError); return ok }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.ok(test.err) {
				t.Fatalf("%T belongs to the wrong error domain", test.err)
			}
		})
	}
}
