package playdate

import "errors"

var (
	ErrVideoUnavailable    = errors.New("video is unavailable")
	ErrVideoPath           = errors.New("video path must not be empty")
	ErrVideoLoad           = errors.New("load video failed")
	ErrVideoClosed         = errors.New("video player is closed")
	ErrVideoFrame          = errors.New("video frame is out of range")
	ErrVideoStreamClosed   = errors.New("video stream is closed")
	ErrVideoBufferSize     = errors.New("video buffer sizes must not be negative")
	ErrVideoSource         = errors.New("video source must be an open runtime file")
	ErrVideoPlayerBorrowed = errors.New("video player is borrowed by a stream")
)

// VideoOperationError reports an error returned by the native video decoder.
type VideoOperationError struct{ Operation, Message string }

func (e VideoOperationError) Error() string {
	if e.Message == "" {
		return e.Operation + " video failed"
	}
	return e.Operation + " video failed: " + e.Message
}
