package playdate

// Filesystem errors.

type fileError string

func (message fileError) Error() string { return string(message) }

// FileOperationError preserves the diagnostic returned by the Playdate API.
type FileOperationError struct {
	Operation string
	Path      string
	Message   string
}

func (failure FileOperationError) Error() string {
	if failure.Message == "" {
		return failure.Operation + " " + failure.Path + ": " + ErrFileIO.Error()
	}
	return failure.Operation + " " + failure.Path + ": " + ErrFileIO.Error() + ": " + failure.Message
}

func (FileOperationError) Unwrap() error { return ErrFileIO }

var (
	// ErrFileClosed indicates an operation on a closed owned file.
	ErrFileClosed error = fileError("file is closed")
	// ErrFilePath indicates a non-relative or parent-traversing path.
	ErrFilePath error = fileError("invalid Playdate file path")
	// ErrFileMode indicates an unsupported combination of open flags.
	ErrFileMode error = fileError("invalid Playdate file mode")
	// ErrFileOffset indicates a seek outside the native signed 32-bit range.
	ErrFileOffset error = fileError("file offset is outside the Playdate range")
	// ErrFileIO categorizes a failure reported by the Playdate filesystem.
	ErrFileIO error = fileError("Playdate filesystem operation failed")
	// ErrFileUnavailable indicates a context without the optional filesystem capability.
	ErrFileUnavailable error = fileError("filesystem capability is unavailable")
)

// Online-scoreboard errors.

type scoreboardError string

func (message scoreboardError) Error() string { return string(message) }

// ScoreboardOperationError preserves a diagnostic returned asynchronously by
// the Playdate scoreboards service.
type ScoreboardOperationError struct {
	Operation string
	BoardID   string
	Message   string
}

func (failure ScoreboardOperationError) Error() string {
	message := failure.Operation
	if failure.BoardID != "" {
		message += " " + failure.BoardID
	}
	if failure.Message != "" {
		message += ": " + failure.Message
	}
	return message
}

func (ScoreboardOperationError) Unwrap() error { return ErrScoreboardRequest }

var (
	ErrScoreboardBoardID     error = scoreboardError("scoreboard board ID is required")
	ErrScoreboardCallback    error = scoreboardError("scoreboard callback is required")
	ErrScoreboardBusy        error = scoreboardError("scoreboard operation is already pending")
	ErrScoreboardRequest     error = scoreboardError("Playdate scoreboard request failed")
	ErrScoreboardUnavailable error = scoreboardError("scoreboards capability is unavailable")
)

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
)

// Graphics errors.

type graphicsError string

func (message graphicsError) Error() string { return string(message) }

var (
	// ErrGraphicsColor indicates a solid color outside the public values.
	ErrGraphicsColor error = graphicsError("invalid graphics color")
	// ErrGraphicsUnavailable indicates a context without the optional graphics slice.
	ErrGraphicsUnavailable error = graphicsError("graphics capability is unavailable")
	// ErrGraphicsGeometry indicates non-positive dimensions, widths, or non-finite angles.
	ErrGraphicsGeometry error = graphicsError("invalid graphics geometry")
	// ErrGraphicsDrawMode indicates a draw mode outside the public values.
	ErrGraphicsDrawMode error = graphicsError("invalid graphics draw mode")
)

// Framebuffer and offscreen drawing errors.

type framebufferError string

func (message framebufferError) Error() string { return string(message) }

type offscreenError string

func (message offscreenError) Error() string { return string(message) }

var (
	// ErrFramebufferExpired indicates use after the scoped callback returned.
	ErrFramebufferExpired error = framebufferError("framebuffer view has expired")
	// ErrFramebufferBounds indicates a pixel or dirty-row range outside the display.
	ErrFramebufferBounds error = framebufferError("framebuffer coordinates are out of range")
	// ErrFramebufferColor indicates a framebuffer pixel color other than black or white.
	ErrFramebufferColor error = framebufferError("framebuffer pixels must be black or white")
	// ErrFramebufferCallback indicates a nil scoped framebuffer callback.
	ErrFramebufferCallback error = framebufferError("framebuffer callback is required")
	// ErrOffscreenCallback indicates a nil offscreen drawing callback.
	ErrOffscreenCallback error = offscreenError("offscreen callback is required")
)

// Tile-map errors.

type tileMapError string

func (message tileMapError) Error() string { return string(message) }

var (
	// ErrTileMapConfig indicates invalid layer dimensions or tile data length.
	ErrTileMapConfig error = tileMapError("invalid tile map configuration")
	// ErrTileMapBitmap indicates that a visible tile has no corresponding bitmap.
	ErrTileMapBitmap error = tileMapError("tile bitmap is unavailable")
	// ErrTileMapDraw indicates invalid drawing inputs or an underlying draw failure.
	ErrTileMapDraw error = tileMapError("tile map draw failed")
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
	// ErrAudioRepeat indicates a repeat count outside the native 32-bit range.
	ErrAudioRepeat error = audioError("audio repeat count must be between zero and 2147483647")
	// ErrAudioRate indicates a zero or non-finite playback rate.
	ErrAudioRate error = audioError("audio rate must be finite and non-zero")
	// ErrAudioReverseUnsupported indicates reverse playback on a streaming player.
	ErrAudioReverseUnsupported error = audioError("streaming audio cannot play in reverse")
	// ErrAudioOffset indicates a non-finite or negative playback offset.
	ErrAudioOffset error = audioError("audio offset must be finite and non-negative")
	// ErrAudioFade indicates a fade duration outside the native signed 32-bit range.
	ErrAudioFade error = audioError("audio fade duration must not exceed 2147483647 frames")
	// ErrAudioUnavailable indicates that an optional advanced audio capability is absent.
	ErrAudioUnavailable error = audioError("advanced audio is unavailable")
	// ErrAudioChannelClosed indicates an operation on a closed routing channel.
	ErrAudioChannelClosed error = audioError("audio channel is closed")
	// ErrAudioSourceInvalid indicates a source not created by this runtime.
	ErrAudioSourceInvalid error = audioError("audio source is invalid")
	// ErrAudioRoute indicates that Playdate rejected a routing graph change.
	ErrAudioRoute error = audioError("audio routing change failed")
	// ErrAudioPan indicates a non-finite pan outside the -1..1 range.
	ErrAudioPan error = audioError("audio pan must be between negative one and one")
	// ErrAudioGraphClosed indicates an operation on a closed synth or signal.
	ErrAudioGraphClosed error = audioError("audio graph node is closed")
	// ErrAudioParameter indicates an invalid non-finite synth or signal value.
	ErrAudioParameter error = audioError("audio parameter must be finite")
	// ErrAudioWaveform indicates a waveform outside the native enumeration.
	ErrAudioWaveform error = audioError("audio waveform is invalid")
	// ErrAudioEventStep indicates a negative control-signal step.
	ErrAudioEventStep error = audioError("audio event step must be non-negative")
)

// Microphone errors.

type microphoneError string

func (message microphoneError) Error() string { return string(message) }

var (
	// ErrMicrophoneUnavailable indicates a context without microphone input.
	ErrMicrophoneUnavailable error = microphoneError("microphone capability is unavailable")
	// ErrMicrophoneCallback indicates a missing permission or recording callback.
	ErrMicrophoneCallback error = microphoneError("microphone callback is required")
	// ErrMicrophoneSource indicates an invalid input-source selection.
	ErrMicrophoneSource error = microphoneError("microphone source is invalid")
	// ErrMicrophoneDenied indicates that microphone permission was denied.
	ErrMicrophoneDenied error = microphoneError("microphone access was denied")
	// ErrMicrophoneStart indicates that native recording could not start.
	ErrMicrophoneStart error = microphoneError("microphone recording failed to start")
	// ErrMicrophoneClosed indicates an operation on a closed recorder.
	ErrMicrophoneClosed error = microphoneError("microphone recorder is closed")
	// ErrMicrophoneSamplesExpired indicates access outside the native callback.
	ErrMicrophoneSamplesExpired error = microphoneError("microphone samples have expired")
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
