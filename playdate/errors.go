package playdate

type bitmapError string

func (message bitmapError) Error() string { return string(message) }

type spriteError string

func (message spriteError) Error() string { return string(message) }

type animationError string

func (message animationError) Error() string { return string(message) }

type audioError string

func (message audioError) Error() string { return string(message) }

// AudioLoadError identifies an audio asset that Playdate could not load.
type AudioLoadError string

func (path AudioLoadError) Error() string { return "load audio: " + string(path) }

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
	// ErrBitmapTableClosed indicates access to a closed bitmap table.
	ErrBitmapTableClosed error = bitmapError("bitmap table is closed")
	// ErrBitmapTableBorrowed indicates an attempt to close a borrowed table.
	ErrBitmapTableBorrowed error = bitmapError("borrowed bitmap table cannot be closed")
	// ErrBitmapFrameRange indicates a negative or missing table frame.
	ErrBitmapFrameRange error = bitmapError("bitmap table frame is out of range")
	// ErrAnimationConfig indicates invalid animation timing or frame bounds.
	ErrAnimationConfig error = animationError("invalid animation configuration")
	// ErrSpriteClosed indicates an operation on an already closed sprite.
	ErrSpriteClosed error = spriteError("sprite is closed")
	// ErrSpriteBorrowed indicates an attempt to close a borrowed query result.
	ErrSpriteBorrowed error = spriteError("borrowed sprite cannot be closed")
	// ErrSpriteCreate indicates that Playdate could not allocate a sprite.
	ErrSpriteCreate error = spriteError("create sprite failed")
	// ErrAudioClosed indicates an operation on a closed audio player.
	ErrAudioClosed error = audioError("audio player is closed")
	// ErrAudioCreate indicates that Playdate could not allocate a player.
	ErrAudioCreate error = audioError("create audio player failed")
	// ErrAudioPlay indicates that Playdate rejected playback.
	ErrAudioPlay error = audioError("audio playback failed")
	// ErrAudioVolume indicates a non-finite volume outside the 0..1 range.
	ErrAudioVolume error = audioError("audio volume must be between zero and one")
)
