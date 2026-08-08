package playdate

// Font errors.

type fontError string

func (message fontError) Error() string { return string(message) }

// FontLoadError contains the diagnostic returned by the Playdate API.
type FontLoadError string

func (message FontLoadError) Error() string { return "load font: " + string(message) }
func (FontLoadError) Is(target error) bool  { return target == ErrFontLoad }

var (
	// ErrFontClosed indicates use of a closed or nil font.
	ErrFontClosed error = fontError("font is closed")
	// ErrFontLoad indicates that a packaged font could not be loaded.
	ErrFontLoad error = fontError("font could not be loaded")
	// ErrFontInvalid indicates a font not created by this runtime.
	ErrFontInvalid error = fontError("font handle was not created by this runtime")
	// ErrFontGlyph indicates that a font does not contain the requested glyph.
	ErrFontGlyph error = fontError("font glyph is unavailable")
)
