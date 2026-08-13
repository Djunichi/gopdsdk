package playdate

// Bitmap errors.

type bitmapError string

func (message bitmapError) Error() string { return string(message) }

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
	// ErrBitmapDataCallback indicates a nil scoped bitmap-data callback.
	ErrBitmapDataCallback error = bitmapError("bitmap data callback is required")
	// ErrBitmapDataExpired indicates use after the bitmap-data callback returned.
	ErrBitmapDataExpired error = bitmapError("bitmap data view has expired")
	// ErrBitmapBounds indicates a pixel outside the bitmap.
	ErrBitmapBounds error = bitmapError("bitmap coordinates are out of range")
	// ErrBitmapMaskSize indicates a mask whose dimensions do not match its bitmap.
	ErrBitmapMaskSize error = bitmapError("bitmap mask dimensions must match")
	// ErrBitmapMask indicates failure to attach a native mask.
	ErrBitmapMask error = bitmapError("set bitmap mask failed")
	// ErrBitmapMaskInUse indicates a bitmap retained by a mask association or view.
	ErrBitmapMaskInUse error = bitmapError("bitmap is still used by a mask")
	// ErrBitmapFlip indicates an unsupported bitmap flip value.
	ErrBitmapFlip error = bitmapError("invalid bitmap flip")
	// ErrBitmapTableSize indicates invalid table dimensions or frame count.
	ErrBitmapTableSize error = bitmapError("bitmap table dimensions and count must be positive")
	// ErrBitmapTableInUse indicates a table retained by a native tilemap.
	ErrBitmapTableInUse error = bitmapError("bitmap table is still used by a tilemap")
	// ErrBitmapMenuImageInUse indicates a bitmap retained as the system menu
	// image. Clear or replace the menu image before closing it.
	ErrBitmapMenuImageInUse error = bitmapError("bitmap is still used as the menu image")
)
