package playdate

// Bitmap errors.

type bitmapError string

func (message bitmapError) Error() string { return string(message) }

// BitmapLoadError contains the diagnostic returned by the Playdate API.
type BitmapLoadError string

func (message BitmapLoadError) Error() string { return "load bitmap: " + string(message) }

var (
	// ErrGraphicsColor indicates a solid color outside the public values.
	ErrGraphicsColor error = bitmapError("invalid graphics color")
	// ErrGraphicsUnavailable indicates a context without the optional graphics slice.
	ErrGraphicsUnavailable error = bitmapError("graphics capability is unavailable")
	// ErrGraphicsGeometry indicates non-positive dimensions, widths, or non-finite angles.
	ErrGraphicsGeometry error = bitmapError("invalid graphics geometry")
	// ErrGraphicsDrawMode indicates a draw mode outside the public values.
	ErrGraphicsDrawMode error = bitmapError("invalid graphics draw mode")
	// ErrBitmapClosed indicates an operation on an already closed bitmap.
	ErrBitmapClosed error = bitmapError("bitmap is closed")
	// ErrBitmapBorrowed indicates an attempt to close a borrowed bitmap.
	ErrBitmapBorrowed error = bitmapError("borrowed bitmap cannot be closed")
	// ErrBitmapColor indicates a color outside the public Color values.
	ErrBitmapColor error = bitmapError("invalid bitmap color")
	// ErrBitmapSize indicates non-positive or unsupported bitmap dimensions.
	ErrBitmapSize error = bitmapError("bitmap dimensions must be positive")
	// ErrBitmapScale indicates a non-positive or non-finite scale.
	ErrBitmapScale error = bitmapError("bitmap scale must be positive and finite")
	// ErrBitmapCreate indicates that Playdate could not allocate a bitmap.
	ErrBitmapCreate error = bitmapError("create bitmap failed")
	// ErrBitmapTableClosed indicates access to a closed bitmap table.
	ErrBitmapTableClosed error = bitmapError("bitmap table is closed")
	// ErrBitmapTableBorrowed indicates an attempt to close a borrowed table.
	ErrBitmapTableBorrowed error = bitmapError("borrowed bitmap table cannot be closed")
	// ErrBitmapFrameRange indicates a negative or missing table frame.
	ErrBitmapFrameRange error = bitmapError("bitmap table frame is out of range")
)

// Animation errors.

type animationError string

func (message animationError) Error() string { return string(message) }

// ErrAnimationConfig indicates invalid animation timing or frame bounds.
var ErrAnimationConfig error = animationError("invalid animation configuration")

// Sprite errors.

type spriteError string

func (message spriteError) Error() string { return string(message) }

var (
	// ErrSpriteClosed indicates an operation on an already closed sprite.
	ErrSpriteClosed error = spriteError("sprite is closed")
	// ErrSpriteBorrowed indicates an attempt to close a borrowed query result.
	ErrSpriteBorrowed error = spriteError("borrowed sprite cannot be closed")
	// ErrSpriteCreate indicates that Playdate could not allocate a sprite.
	ErrSpriteCreate error = spriteError("create sprite failed")
)

// Audio errors.

type audioError string

func (message audioError) Error() string { return string(message) }

// AudioLoadError identifies an audio asset that Playdate could not load.
type AudioLoadError string

func (path AudioLoadError) Error() string { return "load audio: " + string(path) }

var (
	// ErrAudioClosed indicates an operation on a closed audio player.
	ErrAudioClosed error = audioError("audio player is closed")
	// ErrAudioCreate indicates that Playdate could not allocate a player.
	ErrAudioCreate error = audioError("create audio player failed")
	// ErrAudioPlay indicates that Playdate rejected playback.
	ErrAudioPlay error = audioError("audio playback failed")
	// ErrAudioVolume indicates a non-finite volume outside the 0..1 range.
	ErrAudioVolume error = audioError("audio volume must be between zero and one")
)

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
)
