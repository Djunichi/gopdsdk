package playdate

type bitmapError string

func (message bitmapError) Error() string { return string(message) }

type spriteError string

func (message spriteError) Error() string { return string(message) }

// BitmapLoadError contains the diagnostic returned by the Playdate API.
type BitmapLoadError string

func (message BitmapLoadError) Error() string { return "load bitmap: " + string(message) }

var (
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
	// ErrSpriteClosed indicates an operation on an already closed sprite.
	ErrSpriteClosed error = spriteError("sprite is closed")
	// ErrSpriteCreate indicates that Playdate could not allocate a sprite.
	ErrSpriteCreate error = spriteError("create sprite failed")
)
