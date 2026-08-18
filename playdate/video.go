package playdate

// VideoInfo describes a loaded PDV stream and its current decoder position.
type VideoInfo struct {
	Width, Height int
	FrameRate     float32
	FrameCount    int
	CurrentFrame  int
}

// VideoPlayer owns a loaded PDV decoder. A bitmap selected with SetContext is
// borrowed and must remain open until another context is selected or the
// player is closed.
type VideoPlayer interface {
	Info() (VideoInfo, error)
	LastError() (string, error)
	SetContext(Bitmap) error
	UseScreenContext() error
	RenderFrame(frame int) error
	Close() error
}

// Videos is the optional specialized video capability.
type Videos interface {
	LoadVideo(path string) (VideoPlayer, error)
}
