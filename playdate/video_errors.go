package playdate

import "errors"

var (
	ErrVideoUnavailable = errors.New("video is unavailable")
	ErrVideoPath        = errors.New("video path must not be empty")
	ErrVideoLoad        = errors.New("load video failed")
	ErrVideoClosed      = errors.New("video player is closed")
	ErrVideoFrame       = errors.New("video frame is out of range")
)

// VideoOperationError reports an error returned by the native video decoder.
type VideoOperationError struct{ Operation, Message string }

func (e VideoOperationError) Error() string {
	if e.Message == "" {
		return e.Operation + " video failed"
	}
	return e.Operation + " video failed: " + e.Message
}
