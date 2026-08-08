package playdate

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
