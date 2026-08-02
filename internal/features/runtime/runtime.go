// Package runtime coordinates platform-independent Playdate application lifecycle.
package runtime

import (
	"errors"
	"io"
	"math"
	"path"
	"strings"

	"github.com/Djunichi/gopdsdk/playdate"
)

// FileDriver contains platform operations for one owned native file.
type FileDriver struct {
	Read  func(uintptr, []byte) (int, string)
	Write func(uintptr, []byte) (int, string)
	Flush func(uintptr) (int, string)
	Tell  func(uintptr) (int, string)
	Seek  func(uintptr, int32, int) (int, string)
	Close func(uintptr) (int, string)
}

type ownedFile struct {
	handle uintptr
	path   string
	driver FileDriver
	closed bool
}

// NewOwnedFile wraps a Playdate file handle that must be explicitly closed.
func NewOwnedFile(handle uintptr, filePath string, driver FileDriver) playdate.File {
	return &ownedFile{handle: handle, path: filePath, driver: driver}
}

func (file *ownedFile) nativeHandle() (uintptr, error) {
	if file == nil || file.closed || file.handle == 0 {
		return 0, playdate.ErrFileClosed
	}
	return file.handle, nil
}

func fileFailure(operation, filePath, message string) error {
	return playdate.FileOperationError{Operation: operation, Path: filePath, Message: message}
}

func (file *ownedFile) Read(buffer []byte) (int, error) {
	handle, err := file.nativeHandle()
	if err != nil {
		return 0, err
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	count, message := file.driver.Read(handle, buffer)
	if count < 0 {
		return 0, fileFailure("read", file.path, message)
	}
	if count == 0 {
		return 0, io.EOF
	}
	return count, nil
}

func (file *ownedFile) Write(buffer []byte) (int, error) {
	handle, err := file.nativeHandle()
	if err != nil {
		return 0, err
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	count, message := file.driver.Write(handle, buffer)
	if count < 0 {
		return 0, fileFailure("write", file.path, message)
	}
	if count != len(buffer) {
		return count, io.ErrShortWrite
	}
	return count, nil
}

func (file *ownedFile) Flush() error {
	handle, err := file.nativeHandle()
	if err != nil {
		return err
	}
	if result, message := file.driver.Flush(handle); result < 0 {
		return fileFailure("flush", file.path, message)
	}
	return nil
}

func (file *ownedFile) Seek(offset int64, whence int) (int64, error) {
	handle, err := file.nativeHandle()
	if err != nil {
		return 0, err
	}
	if offset < math.MinInt32 || offset > math.MaxInt32 || whence < io.SeekStart || whence > io.SeekEnd {
		return 0, playdate.ErrFileOffset
	}
	if result, message := file.driver.Seek(handle, int32(offset), whence); result < 0 {
		return 0, fileFailure("seek", file.path, message)
	}
	position, message := file.driver.Tell(handle)
	if position < 0 {
		return 0, fileFailure("tell", file.path, message)
	}
	return int64(position), nil
}

func (file *ownedFile) Close() error {
	handle, err := file.nativeHandle()
	if err != nil {
		return err
	}
	file.closed = true
	file.handle = 0
	if result, message := file.driver.Close(handle); result < 0 {
		return fileFailure("close", file.path, message)
	}
	return nil
}

// ValidateFilePath applies the shared Playdate relative-path contract.
func ValidateFilePath(filePath string, allowRoot bool) error {
	if strings.IndexByte(filePath, 0) >= 0 || strings.Contains(filePath, "\\") || strings.HasPrefix(filePath, "/") {
		return playdate.ErrFilePath
	}
	cleaned := path.Clean(filePath)
	if cleaned == "." {
		if allowRoot {
			return nil
		}
		return playdate.ErrFilePath
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || cleaned != filePath {
		return playdate.ErrFilePath
	}
	return nil
}

// ValidateFileOptions applies the official FileOptions combinations.
func ValidateFileOptions(options playdate.FileOptions) error {
	switch options {
	case playdate.FileReadPackage, playdate.FileReadData,
		playdate.FileReadPackage | playdate.FileReadData,
		playdate.FileWrite, playdate.FileAppend:
		return nil
	default:
		return playdate.ErrFileMode
	}
}

// Framebuffer provides a direct, callback-scoped view of native display memory.
type Framebuffer struct {
	data                 []byte
	width, height        int
	rowBytes             int
	active               bool
	dirtyStart, dirtyEnd int
}

// WithFramebuffer scopes a native framebuffer view and reports its combined
// dirty row range after successful or failed callback execution.
func WithFramebuffer(data []byte, width, height, rowBytes int, mark func(int, int), callback func(playdate.Framebuffer) error) error {
	if callback == nil {
		return playdate.ErrFramebufferCallback
	}
	frame := &Framebuffer{data: data, width: width, height: height, rowBytes: rowBytes, active: true, dirtyStart: height, dirtyEnd: -1}
	err := callback(frame)
	frame.active = false
	frame.data = nil
	if frame.dirtyEnd >= frame.dirtyStart && mark != nil {
		mark(frame.dirtyStart, frame.dirtyEnd)
	}
	return err
}

func (f *Framebuffer) valid() error {
	if f == nil || !f.active {
		return playdate.ErrFramebufferExpired
	}
	return nil
}

func (f *Framebuffer) Width() int    { return f.width }
func (f *Framebuffer) Height() int   { return f.height }
func (f *Framebuffer) RowBytes() int { return f.rowBytes }
func (f *Framebuffer) Bytes() ([]byte, error) {
	if err := f.valid(); err != nil {
		return nil, err
	}
	return f.data, nil
}
func (f *Framebuffer) Pixel(x, y int) (playdate.Color, error) {
	if err := f.valid(); err != nil {
		return 0, err
	}
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return 0, playdate.ErrFramebufferBounds
	}
	if f.data[y*f.rowBytes+x/8]&(0x80>>uint(x&7)) != 0 {
		return playdate.ColorWhite, nil
	}
	return playdate.ColorBlack, nil
}
func (f *Framebuffer) SetPixel(x, y int, color playdate.Color) error {
	if err := f.valid(); err != nil {
		return err
	}
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return playdate.ErrFramebufferBounds
	}
	if color != playdate.ColorBlack && color != playdate.ColorWhite {
		return playdate.ErrFramebufferColor
	}
	index, mask := y*f.rowBytes+x/8, byte(0x80>>uint(x&7))
	if color == playdate.ColorWhite {
		f.data[index] |= mask
	} else {
		f.data[index] &^= mask
	}
	return f.MarkDirtyRows(y, y)
}
func (f *Framebuffer) MarkDirtyRows(start, end int) error {
	if err := f.valid(); err != nil {
		return err
	}
	if start < 0 || end < start || end >= f.height {
		return playdate.ErrFramebufferBounds
	}
	if start < f.dirtyStart {
		f.dirtyStart = start
	}
	if end > f.dirtyEnd {
		f.dirtyEnd = end
	}
	return nil
}

// ValidatePrimitiveGeometry applies the shared primitive dimension contract.
func ValidatePrimitiveGeometry(width, height, lineWidth int, startAngle, endAngle float32) error {
	if width <= 0 || height <= 0 || lineWidth <= 0 || math.IsNaN(float64(startAngle)) || math.IsNaN(float64(endAngle)) || math.IsInf(float64(startAngle), 0) || math.IsInf(float64(endAngle), 0) {
		return playdate.ErrGraphicsGeometry
	}
	return nil
}

// ValidateDrawMode applies the shared draw-mode contract.
func ValidateDrawMode(mode playdate.DrawMode) error {
	if mode > playdate.DrawModeInverted {
		return playdate.ErrGraphicsDrawMode
	}
	return nil
}

// Event identifies an event delivered by the Playdate runtime.
type Event int32

// Event values mirror PDSystemEvent in the official Playdate C API.
const (
	EventInit Event = iota
	EventInitLua
	EventLock
	EventUnlock
	EventPause
	EventResume
	EventTerminate
	EventKeyPressed
	EventKeyReleased
	EventLowPower
	EventMirrorStarted
	EventMirrorEnded
)

var (
	// ErrInitRequired indicates a missing application initialization callback.
	ErrInitRequired = errors.New("runtime init callback is required")
	// ErrUpdateRequired indicates a missing application update callback.
	ErrUpdateRequired = errors.New("runtime update callback is required")
	// ErrAlreadyInitialized indicates a duplicate initialization event.
	ErrAlreadyInitialized = errors.New("runtime is already initialized")
	// ErrNotInitialized indicates an update before successful initialization.
	ErrNotInitialized = errors.New("runtime is not initialized")
	// ErrFailed indicates that an earlier callback permanently stopped runtime.
	ErrFailed = errors.New("runtime callback previously failed")
	// ErrTerminated indicates a callback after successful termination.
	ErrTerminated    = errors.New("runtime is terminated")
	ErrInvalidBitmap = errors.New("bitmap handle was not created by this runtime")
)

// FontDriver contains platform operations for an owned native font.
type FontDriver struct {
	TextWidth func(uintptr, string) int
	Height    func(uintptr) int
	Free      func(uintptr)
}

type font struct {
	handle uintptr
	driver FontDriver
	closed bool
}

// NewOwnedFont wraps a font returned by the Playdate loader.
func NewOwnedFont(handle uintptr, driver FontDriver) playdate.Font {
	return &font{handle: handle, driver: driver}
}

func (f *font) nativeHandle() (uintptr, error) {
	if f == nil || f.closed || f.handle == 0 {
		return 0, playdate.ErrFontClosed
	}
	return f.handle, nil
}

func (f *font) TextWidth(text string) (int, error) {
	handle, err := f.nativeHandle()
	if err != nil {
		return 0, err
	}
	return f.driver.TextWidth(handle, text), nil
}

func (f *font) Height() (int, error) {
	handle, err := f.nativeHandle()
	if err != nil {
		return 0, err
	}
	return f.driver.Height(handle), nil
}

func (f *font) Close() error {
	handle, err := f.nativeHandle()
	if err != nil {
		return err
	}
	f.driver.Free(handle)
	f.handle = 0
	f.closed = true
	return nil
}

// FontHandle validates and extracts a font created by this runtime.
func FontHandle(value playdate.Font) (uintptr, error) {
	f, ok := value.(*font)
	if !ok {
		return 0, playdate.ErrFontInvalid
	}
	return f.nativeHandle()
}

// AudioDriver contains the common native operations for either accepted player.
type AudioDriver struct {
	Play      func(uintptr) bool
	Stop      func(uintptr)
	SetVolume func(uintptr, float32, float32)
	Volume    func(uintptr) (float32, float32)
	IsPlaying func(uintptr) bool
	Pause     func(uintptr, bool)
	Free      func(uintptr)
}

type audioPlayer struct {
	handle uintptr
	driver AudioDriver
	paused bool
	closed bool
}

// NewSoundEffect wraps an owned sample player and its sample as one handle.
func NewSoundEffect(handle uintptr, driver AudioDriver) playdate.SoundEffect {
	return &audioPlayer{handle: handle, driver: driver}
}

// NewFilePlayer wraps an owned streaming file player.
func NewFilePlayer(handle uintptr, driver AudioDriver) playdate.FilePlayer {
	return &audioPlayer{handle: handle, driver: driver}
}

func (p *audioPlayer) nativeHandle() (uintptr, error) {
	if p == nil || p.closed || p.handle == 0 {
		return 0, playdate.ErrAudioClosed
	}
	return p.handle, nil
}

func (p *audioPlayer) Play() error {
	handle, err := p.nativeHandle()
	if err != nil {
		return err
	}
	if !p.driver.Play(handle) {
		return playdate.ErrAudioPlay
	}
	p.paused = false
	return nil
}

func (p *audioPlayer) Stop() error {
	handle, err := p.nativeHandle()
	if err != nil {
		return err
	}
	p.driver.Stop(handle)
	p.paused = false
	return nil
}

func (p *audioPlayer) SetVolume(left, right float32) error {
	if err := ValidateAudioVolume(left, right); err != nil {
		return err
	}
	handle, err := p.nativeHandle()
	if err != nil {
		return err
	}
	p.driver.SetVolume(handle, left, right)
	return nil
}

func (p *audioPlayer) Volume() (float32, float32, error) {
	handle, err := p.nativeHandle()
	if err != nil {
		return 0, 0, err
	}
	left, right := p.driver.Volume(handle)
	return left, right, nil
}

func (p *audioPlayer) State() (playdate.PlaybackState, error) {
	handle, err := p.nativeHandle()
	if err != nil {
		return 0, err
	}
	if p.paused {
		return playdate.PlaybackPaused, nil
	}
	if p.driver.IsPlaying(handle) {
		return playdate.PlaybackPlaying, nil
	}
	return playdate.PlaybackStopped, nil
}

func (p *audioPlayer) Pause() error {
	handle, err := p.nativeHandle()
	if err != nil {
		return err
	}
	if p.driver.IsPlaying(handle) {
		p.driver.Pause(handle, true)
		p.paused = true
	}
	return nil
}

func (p *audioPlayer) Resume() error {
	handle, err := p.nativeHandle()
	if err != nil {
		return err
	}
	if p.paused {
		p.driver.Pause(handle, false)
		p.paused = false
	}
	return nil
}

func (p *audioPlayer) Close() error {
	handle, err := p.nativeHandle()
	if err != nil {
		return err
	}
	p.driver.Stop(handle)
	p.driver.Free(handle)
	p.handle = 0
	p.closed = true
	p.paused = false
	return nil
}

// ValidateAudioVolume applies the shared Simulator/device volume contract.
func ValidateAudioVolume(left, right float32) error {
	if left < 0 || left > 1 || right < 0 || right > 1 || left != left || right != right {
		return playdate.ErrAudioVolume
	}
	return nil
}

// BitmapDriver contains platform operations for one native bitmap handle.
type BitmapDriver struct {
	Dimensions func(uintptr) (width, height int)
	Fill       func(uintptr, playdate.Color)
	Free       func(uintptr)
}

// Bitmap owns or borrows a native Playdate bitmap.
type Bitmap struct {
	handle uintptr
	driver BitmapDriver
	table  *BitmapTable
	owned  bool
	closed bool
}

// BitmapTableDriver contains platform operations for one native bitmap table.
type BitmapTableDriver struct {
	Frame func(table uintptr, index int) uintptr
	Free  func(uintptr)
}

// BitmapTable owns or borrows a native Playdate bitmap table.
type BitmapTable struct {
	handle        uintptr
	driver        BitmapTableDriver
	bitmapDriver  BitmapDriver
	owned, closed bool
}

// NewOwnedBitmapTable wraps a table and its borrowed frames.
func NewOwnedBitmapTable(handle uintptr, driver BitmapTableDriver, bitmapDriver BitmapDriver) *BitmapTable {
	return &BitmapTable{handle: handle, driver: driver, bitmapDriver: bitmapDriver, owned: true}
}

// NewBorrowedBitmapTable wraps a table controlled by Playdate.
func NewBorrowedBitmapTable(handle uintptr, driver BitmapTableDriver, bitmapDriver BitmapDriver) *BitmapTable {
	return &BitmapTable{handle: handle, driver: driver, bitmapDriver: bitmapDriver}
}

func (t *BitmapTable) Frame(index int) (playdate.Bitmap, error) {
	if t == nil || t.closed || t.handle == 0 {
		return nil, playdate.ErrBitmapTableClosed
	}
	if index < 0 {
		return nil, playdate.ErrBitmapFrameRange
	}
	handle := t.driver.Frame(t.handle, index)
	if handle == 0 {
		return nil, playdate.ErrBitmapFrameRange
	}
	return &Bitmap{handle: handle, driver: t.bitmapDriver, table: t}, nil
}

func (t *BitmapTable) Close() error {
	if t == nil || t.closed || t.handle == 0 {
		return playdate.ErrBitmapTableClosed
	}
	if !t.owned {
		return playdate.ErrBitmapTableBorrowed
	}
	t.driver.Free(t.handle)
	t.closed = true
	t.handle = 0
	return nil
}

// NewOwnedBitmap wraps a bitmap that must be explicitly closed.
func NewOwnedBitmap(handle uintptr, driver BitmapDriver) *Bitmap {
	return &Bitmap{handle: handle, driver: driver, owned: true}
}

// NewBorrowedBitmap wraps a bitmap whose lifetime is controlled by Playdate.
func NewBorrowedBitmap(handle uintptr, driver BitmapDriver) *Bitmap {
	return &Bitmap{handle: handle, driver: driver}
}

func (b *Bitmap) nativeHandle() (uintptr, error) {
	if b == nil || b.closed || b.handle == 0 || b.table != nil && b.table.closed {
		return 0, playdate.ErrBitmapClosed
	}
	return b.handle, nil
}

func (b *Bitmap) Width() (int, error)  { width, _, err := b.dimensions(); return width, err }
func (b *Bitmap) Height() (int, error) { _, height, err := b.dimensions(); return height, err }
func (b *Bitmap) dimensions() (int, int, error) {
	handle, err := b.nativeHandle()
	if err != nil {
		return 0, 0, err
	}
	w, h := b.driver.Dimensions(handle)
	return w, h, nil
}
func (b *Bitmap) Clear() error { return b.Fill(playdate.ColorClear) }
func (b *Bitmap) Fill(color playdate.Color) error {
	handle, err := b.nativeHandle()
	if err != nil {
		return err
	}
	if color > playdate.ColorBlack {
		return playdate.ErrBitmapColor
	}
	b.driver.Fill(handle, color)
	return nil
}
func (b *Bitmap) Close() error {
	if b == nil || b.closed || b.handle == 0 {
		return playdate.ErrBitmapClosed
	}
	if !b.owned {
		return playdate.ErrBitmapBorrowed
	}
	b.driver.Free(b.handle)
	b.closed = true
	b.handle = 0
	return nil
}

// BitmapHandle validates and extracts a runtime bitmap for platform drawing.
func BitmapHandle(bitmap playdate.Bitmap) (uintptr, error) {
	value, ok := bitmap.(*Bitmap)
	if !ok {
		return 0, ErrInvalidBitmap
	}
	return value.nativeHandle()
}

// OwnedBitmapHandle validates a bitmap suitable for an offscreen context.
func OwnedBitmapHandle(bitmap playdate.Bitmap) (uintptr, error) {
	value, ok := bitmap.(*Bitmap)
	if !ok {
		return 0, ErrInvalidBitmap
	}
	if !value.owned {
		return 0, playdate.ErrBitmapBorrowed
	}
	return value.nativeHandle()
}

// SpriteDriver contains platform operations for one native sprite handle.
type SpriteDriver struct {
	SetBitmap          func(sprite, bitmap uintptr)
	MoveTo             func(uintptr, float32, float32)
	MoveBy             func(uintptr, float32, float32)
	SetVisible         func(uintptr, bool)
	SetZIndex          func(uintptr, int)
	SetCollideRect     func(uintptr, playdate.Rect)
	ClearCollideRect   func(uintptr)
	SetTag             func(uintptr, uint8)
	MoveWithCollisions func(uintptr, float32, float32) (float32, float32, []NativeCollision)
	Add                func(uintptr)
	Remove             func(uintptr)
	Free               func(uintptr)
}

// NativeCollision is the adapter-facing representation of SpriteCollisionInfo.
type NativeCollision struct {
	Other                 uintptr
	ResponseType          playdate.CollisionResponse
	Overlaps              bool
	Time                  float32
	Move, Normal, Touch   playdate.Point
	SpriteRect, OtherRect playdate.Rect
}

// Sprite owns a native Playdate sprite.
type Sprite struct {
	handle uintptr
	driver SpriteDriver
	added  bool
	closed bool
	owned  bool
}

// NewOwnedSprite wraps a sprite that must be explicitly closed.
func NewOwnedSprite(handle uintptr, driver SpriteDriver) *Sprite {
	return &Sprite{handle: handle, driver: driver, owned: true}
}

// NewBorrowedSprite wraps a query result owned by the Playdate display list.
func NewBorrowedSprite(handle uintptr, driver SpriteDriver) *Sprite {
	return &Sprite{handle: handle, driver: driver}
}

// SpriteHandle validates and extracts a runtime sprite for platform queries.
func SpriteHandle(sprite playdate.Sprite) (uintptr, error) {
	value, ok := sprite.(*Sprite)
	if !ok {
		return 0, playdate.ErrSpriteClosed
	}
	return value.nativeHandle()
}

func (s *Sprite) nativeHandle() (uintptr, error) {
	if s == nil || s.closed || s.handle == 0 {
		return 0, playdate.ErrSpriteClosed
	}
	return s.handle, nil
}

func (s *Sprite) SetBitmap(bitmap playdate.Bitmap) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	bitmapHandle, err := BitmapHandle(bitmap)
	if err != nil {
		return err
	}
	s.driver.SetBitmap(handle, bitmapHandle)
	return nil
}
func (s *Sprite) SetPosition(x, y float32) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.MoveTo(handle, x, y)
	return nil
}
func (s *Sprite) MoveBy(dx, dy float32) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.MoveBy(handle, dx, dy)
	return nil
}
func (s *Sprite) SetVisible(visible bool) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.SetVisible(handle, visible)
	return nil
}
func (s *Sprite) SetZIndex(z int) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.SetZIndex(handle, z)
	return nil
}
func (s *Sprite) SetCollideRect(rect playdate.Rect) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.SetCollideRect(handle, rect)
	return nil
}
func (s *Sprite) ClearCollideRect() error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.ClearCollideRect(handle)
	return nil
}
func (s *Sprite) SetTag(tag uint8) error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	s.driver.SetTag(handle, tag)
	return nil
}
func (s *Sprite) MoveWithCollisions(x, y float32) (playdate.MoveResult, error) {
	handle, err := s.nativeHandle()
	if err != nil {
		return playdate.MoveResult{}, err
	}
	actualX, actualY, native := s.driver.MoveWithCollisions(handle, x, y)
	result := playdate.MoveResult{ActualX: actualX, ActualY: actualY, Collisions: make([]playdate.Collision, len(native))}
	for index, collision := range native {
		result.Collisions[index] = playdate.Collision{Other: NewBorrowedSprite(collision.Other, s.driver), ResponseType: collision.ResponseType, Overlaps: collision.Overlaps, Time: collision.Time, Move: collision.Move, Normal: collision.Normal, Touch: collision.Touch, SpriteRect: collision.SpriteRect, OtherRect: collision.OtherRect}
	}
	return result, nil
}
func (s *Sprite) Add() error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	if !s.added {
		s.driver.Add(handle)
		s.added = true
	}
	return nil
}
func (s *Sprite) Remove() error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	if s.added {
		s.driver.Remove(handle)
		s.added = false
	}
	return nil
}
func (s *Sprite) Close() error {
	handle, err := s.nativeHandle()
	if err != nil {
		return err
	}
	if !s.owned {
		return playdate.ErrSpriteBorrowed
	}
	if s.added {
		s.driver.Remove(handle)
		s.added = false
	}
	s.driver.Free(handle)
	s.closed = true
	s.handle = 0
	return nil
}

// BorrowedSprites converts native query handles without transferring ownership.
func BorrowedSprites(handles []uintptr, driver SpriteDriver) []playdate.Sprite {
	result := make([]playdate.Sprite, len(handles))
	for index, handle := range handles {
		result[index] = NewBorrowedSprite(handle, driver)
	}
	return result
}

// ValidateBitmapSize applies the common Simulator/device creation contract.
func ValidateBitmapSize(width, height int) error {
	if width <= 0 || height <= 0 || width > 2147483647 || height > 2147483647 {
		return playdate.ErrBitmapSize
	}
	return nil
}

// ValidateBitmapScale applies the common Simulator/device drawing contract.
func ValidateBitmapScale(scaleX, scaleY float32) error {
	const maxFloat32 = 3.4028235e38
	if scaleX <= 0 || scaleY <= 0 || scaleX != scaleX || scaleY != scaleY || scaleX > maxFloat32 || scaleY > maxFloat32 {
		return playdate.ErrBitmapScale
	}
	return nil
}

// RawInput contains one platform input sample. Both platform adapters provide
// this same representation before the runtime derives frame transitions.
type RawInput struct {
	Buttons      playdate.Buttons
	CrankAngle   float32
	CrankDelta   float32
	CrankDocked  bool
	DeltaSeconds float32
}

// Callbacks contains application behavior invoked by Runtime.
type Callbacks struct {
	Init      func() error
	Lifecycle func(playdate.LifecycleEvent) error
	Update    func(playdate.Input) (refresh bool, err error)
}

// Runtime enforces the platform-independent application lifecycle.
type Runtime struct {
	callbacks   Callbacks
	initialized bool
	failed      bool
	terminated  bool
	input       playdate.Input
	hasInput    bool
}

// Application is the platform-independent entry point shared by Simulator and
// device ABI adapters.
type Application struct {
	runtime *Runtime
}

type applicationContext struct {
	playdate.Context
	input                playdate.Input
	menuItems            []playdate.MenuItem
	accelerometerEnabled bool
}

func (context *applicationContext) Input() playdate.Input { return context.input }

func (context *applicationContext) ExitToLauncher() {
	launcher, ok := context.Context.(playdate.Launcher)
	if ok {
		launcher.ExitToLauncher()
	}
}

func (context *applicationContext) SetAccelerometerEnabled(enabled bool) {
	accelerometer, ok := context.Context.(playdate.Accelerometer)
	if !ok {
		return
	}
	accelerometer.SetAccelerometerEnabled(enabled)
	context.accelerometerEnabled = enabled
}

func (context *applicationContext) AccelerometerXYZ() (x, y, z float32) {
	accelerometer, ok := context.Context.(playdate.Accelerometer)
	if !ok || !context.accelerometerEnabled {
		return 0, 0, 0
	}
	return accelerometer.AccelerometerXYZ()
}

func (context *applicationContext) PowerStatus() playdate.PowerStatus {
	monitor, _ := context.Context.(playdate.PowerMonitor)
	if monitor == nil {
		return 0
	}
	return monitor.PowerStatus()
}
func (context *applicationContext) BatteryPercentage() float32 {
	monitor, _ := context.Context.(playdate.PowerMonitor)
	if monitor == nil {
		return 0
	}
	return monitor.BatteryPercentage()
}
func (context *applicationContext) BatteryVoltage() float32 {
	monitor, _ := context.Context.(playdate.PowerMonitor)
	if monitor == nil {
		return 0
	}
	return monitor.BatteryVoltage()
}
func (context *applicationContext) SystemVolume() float32 {
	settings, _ := context.Context.(playdate.SystemPreferences)
	if settings == nil {
		return 0
	}
	return settings.SystemVolume()
}
func (context *applicationContext) ReduceFlashing() bool {
	settings, _ := context.Context.(playdate.SystemPreferences)
	return settings != nil && settings.ReduceFlashing()
}
func (context *applicationContext) TimezoneOffsetSeconds() int32 {
	settings, _ := context.Context.(playdate.SystemPreferences)
	if settings == nil {
		return 0
	}
	return settings.TimezoneOffsetSeconds()
}
func (context *applicationContext) Uses24HourTime() bool {
	settings, _ := context.Context.(playdate.SystemPreferences)
	return settings != nil && settings.Uses24HourTime()
}

func (context *applicationContext) disableAccelerometer() {
	if context.accelerometerEnabled {
		context.SetAccelerometerEnabled(false)
	}
}

func (context *applicationContext) systemMenu() (playdate.SystemMenu, error) {
	menu, ok := context.Context.(playdate.SystemMenu)
	if !ok {
		return nil, playdate.ErrMenuItemCreate
	}
	return menu, nil
}

func (context *applicationContext) AddActionMenuItem(title string, callback func()) (playdate.MenuItem, error) {
	menu, err := context.systemMenu()
	if err != nil {
		return nil, err
	}
	item, err := menu.AddActionMenuItem(title, callback)
	if err == nil {
		context.menuItems = append(context.menuItems, item)
	}
	return item, err
}

func (context *applicationContext) AddCheckmarkMenuItem(title string, value bool, callback func()) (playdate.CheckmarkMenuItem, error) {
	menu, err := context.systemMenu()
	if err != nil {
		return nil, err
	}
	item, err := menu.AddCheckmarkMenuItem(title, value, callback)
	if err == nil {
		context.menuItems = append(context.menuItems, item)
	}
	return item, err
}

func (context *applicationContext) AddOptionsMenuItem(title string, options []string, callback func()) (playdate.OptionsMenuItem, error) {
	menu, err := context.systemMenu()
	if err != nil {
		return nil, err
	}
	item, err := menu.AddOptionsMenuItem(title, options, callback)
	if err == nil {
		context.menuItems = append(context.menuItems, item)
	}
	return item, err
}

func (context *applicationContext) Language() playdate.Language {
	localization, ok := context.Context.(playdate.Localization)
	if !ok {
		return playdate.LanguageEnglish
	}
	return localization.Language()
}

func (context *applicationContext) LocalizedText(key string, language playdate.Language) (string, bool) {
	localization, ok := context.Context.(playdate.Localization)
	if !ok {
		return "", false
	}
	return localization.LocalizedText(key, language)
}

func (context *applicationContext) removeMenuItems() {
	for _, item := range context.menuItems {
		item.Remove()
	}
	context.menuItems = nil
}

func (context *applicationContext) fileSystem() (playdate.FileSystem, error) {
	files, ok := context.Context.(playdate.FileSystem)
	if !ok {
		return nil, playdate.ErrFileUnavailable
	}
	return files, nil
}

func (context *applicationContext) OpenFile(filePath string, options playdate.FileOptions) (playdate.File, error) {
	files, err := context.fileSystem()
	if err != nil {
		return nil, err
	}
	return files.OpenFile(filePath, options)
}

func (context *applicationContext) Stat(filePath string) (playdate.FileInfo, error) {
	files, err := context.fileSystem()
	if err != nil {
		return playdate.FileInfo{}, err
	}
	return files.Stat(filePath)
}

func (context *applicationContext) List(filePath string, showHidden bool) ([]string, error) {
	files, err := context.fileSystem()
	if err != nil {
		return nil, err
	}
	return files.List(filePath, showHidden)
}

func (context *applicationContext) Mkdir(filePath string) error {
	files, err := context.fileSystem()
	if err != nil {
		return err
	}
	return files.Mkdir(filePath)
}

func (context *applicationContext) Remove(filePath string, recursive bool) error {
	files, err := context.fileSystem()
	if err != nil {
		return err
	}
	return files.Remove(filePath, recursive)
}

func (context *applicationContext) Rename(from, to string) error {
	files, err := context.fileSystem()
	if err != nil {
		return err
	}
	return files.Rename(from, to)
}

func (context *applicationContext) LoadFont(path string) (playdate.Font, error) {
	graphics, ok := context.Context.(playdate.FontGraphics)
	if !ok {
		return nil, playdate.ErrFontLoad
	}
	return graphics.LoadFont(path)
}

func (context *applicationContext) DrawTextFont(font playdate.Font, text string, x, y int) error {
	graphics, ok := context.Context.(playdate.FontGraphics)
	if !ok {
		return playdate.ErrFontLoad
	}
	return graphics.DrawTextFont(font, text, x, y)
}

func (context *applicationContext) primitiveGraphics() (playdate.PrimitiveGraphics, error) {
	graphics, ok := context.Context.(playdate.PrimitiveGraphics)
	if !ok {
		return nil, playdate.ErrGraphicsUnavailable
	}
	return graphics, nil
}

func (context *applicationContext) DrawLine(x1, y1, x2, y2, width int, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.DrawLine(x1, y1, x2, y2, width, paint)
}
func (context *applicationContext) DrawRect(x, y, width, height int, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.DrawRect(x, y, width, height, paint)
}
func (context *applicationContext) FillRect(x, y, width, height int, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.FillRect(x, y, width, height, paint)
}
func (context *applicationContext) DrawEllipse(x, y, width, height, lineWidth int, startAngle, endAngle float32, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.DrawEllipse(x, y, width, height, lineWidth, startAngle, endAngle, paint)
}
func (context *applicationContext) FillEllipse(x, y, width, height int, startAngle, endAngle float32, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.FillEllipse(x, y, width, height, startAngle, endAngle, paint)
}
func (context *applicationContext) DrawTriangle(x1, y1, x2, y2, x3, y3, width int, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.DrawTriangle(x1, y1, x2, y2, x3, y3, width, paint)
}
func (context *applicationContext) FillTriangle(x1, y1, x2, y2, x3, y3 int, paint playdate.Paint) error {
	graphics, err := context.primitiveGraphics()
	if err != nil {
		return err
	}
	return graphics.FillTriangle(x1, y1, x2, y2, x3, y3, paint)
}

func (context *applicationContext) graphicsState() (playdate.GraphicsState, error) {
	graphics, ok := context.Context.(playdate.GraphicsState)
	if !ok {
		return nil, playdate.ErrGraphicsUnavailable
	}
	return graphics, nil
}
func (context *applicationContext) SetClipRect(x, y, width, height int) error {
	graphics, err := context.graphicsState()
	if err != nil {
		return err
	}
	return graphics.SetClipRect(x, y, width, height)
}
func (context *applicationContext) ClearClipRect() {
	if graphics, err := context.graphicsState(); err == nil {
		graphics.ClearClipRect()
	}
}
func (context *applicationContext) SetDrawOffset(dx, dy int) {
	if graphics, err := context.graphicsState(); err == nil {
		graphics.SetDrawOffset(dx, dy)
	}
}
func (context *applicationContext) SetDrawMode(mode playdate.DrawMode) error {
	graphics, err := context.graphicsState()
	if err != nil {
		return err
	}
	return graphics.SetDrawMode(mode)
}

func (context *applicationContext) WithFramebuffer(callback func(playdate.Framebuffer) error) error {
	graphics, ok := context.Context.(playdate.FramebufferGraphics)
	if !ok {
		return playdate.ErrGraphicsUnavailable
	}
	return graphics.WithFramebuffer(callback)
}

func (context *applicationContext) DrawInto(bitmap playdate.Bitmap, callback func() error) error {
	graphics, ok := context.Context.(playdate.OffscreenGraphics)
	if !ok {
		return playdate.ErrGraphicsUnavailable
	}
	return graphics.DrawInto(bitmap, callback)
}

// NewApplication composes a public game with its platform context. beforeInit
// runs immediately before the game's initialization callback when a platform
// adapter needs to prepare callback state.
func NewApplication(game playdate.Game, context playdate.Context, beforeInit func()) (*Application, error) {
	gameContext := &applicationContext{Context: context}
	lifecycle, _ := game.(playdate.LifecycleGame)
	runtime, err := New(Callbacks{
		Init: func() error {
			if beforeInit != nil {
				beforeInit()
			}
			return game.Init(gameContext)
		},
		Lifecycle: func(event playdate.LifecycleEvent) error {
			var err error
			if lifecycle != nil {
				err = lifecycle.HandleLifecycle(gameContext, event)
			}
			if event == playdate.LifecycleTerminate {
				gameContext.disableAccelerometer()
				gameContext.removeMenuItems()
			}
			return err
		},
		Update: func(input playdate.Input) (bool, error) {
			gameContext.input = input
			return game.Update(gameContext)
		},
	})
	if err != nil {
		return nil, err
	}
	return &Application{runtime: runtime}, nil
}

// Handle delivers a Playdate system event to the application lifecycle.
func (a *Application) Handle(event Event, arg uint32) error {
	return a.runtime.Handle(event, arg)
}

// Update invokes the application's frame callback.
func (a *Application) Update(input RawInput) (int32, error) {
	return a.runtime.Update(input)
}

// New validates callbacks and returns an uninitialized runtime.
func New(callbacks Callbacks) (*Runtime, error) {
	if callbacks.Init == nil {
		return nil, ErrInitRequired
	}
	if callbacks.Update == nil {
		return nil, ErrUpdateRequired
	}
	if callbacks.Lifecycle == nil {
		callbacks.Lifecycle = func(playdate.LifecycleEvent) error { return nil }
	}
	return &Runtime{callbacks: callbacks}, nil
}

// Handle delivers a Playdate system event to the lifecycle.
func (r *Runtime) Handle(event Event, _ uint32) error {
	if r.failed {
		return ErrFailed
	}
	if r.terminated {
		return ErrTerminated
	}
	if event == EventInit {
		if r.initialized {
			return ErrAlreadyInitialized
		}
		if err := r.callbacks.Init(); err != nil {
			r.failed = true
			return err
		}
		r.initialized = true
		return nil
	}
	if !r.initialized {
		return ErrNotInitialized
	}
	lifecycleEvent, ok := lifecycleEvent(event)
	if !ok {
		return nil
	}
	if err := r.callbacks.Lifecycle(lifecycleEvent); err != nil {
		r.failed = true
		return err
	}
	if event == EventTerminate {
		r.terminated = true
	}
	return nil
}

func lifecycleEvent(event Event) (playdate.LifecycleEvent, bool) {
	switch event {
	case EventPause:
		return playdate.LifecyclePause, true
	case EventResume:
		return playdate.LifecycleResume, true
	case EventLock:
		return playdate.LifecycleLock, true
	case EventUnlock:
		return playdate.LifecycleUnlock, true
	case EventTerminate:
		return playdate.LifecycleTerminate, true
	case EventLowPower:
		return playdate.LifecycleLowPower, true
	default:
		return 0, false
	}
}

// Update derives and delivers the next frame's portable input snapshot.
func (r *Runtime) Update(raw RawInput) (int32, error) {
	if r.failed {
		return 0, ErrFailed
	}
	if r.terminated {
		return 0, ErrTerminated
	}
	if !r.initialized {
		return 0, ErrNotInitialized
	}
	previousButtons := r.input.Buttons
	previousDocked := raw.CrankDocked
	if r.hasInput {
		previousDocked = r.input.CrankDocked
	}
	r.input = playdate.Input{
		Buttons: raw.Buttons, Pressed: raw.Buttons &^ previousButtons,
		Released: previousButtons &^ raw.Buttons, Held: raw.Buttons & previousButtons,
		CrankAngle: raw.CrankAngle, CrankDelta: raw.CrankDelta,
		CrankDocked: raw.CrankDocked, CrankDockedThisFrame: r.hasInput && raw.CrankDocked && !previousDocked,
		CrankUndocked: r.hasInput && !raw.CrankDocked && previousDocked,
		DeltaSeconds:  raw.DeltaSeconds,
	}
	r.hasInput = true
	shouldRefresh, err := r.callbacks.Update(r.input)
	if err != nil {
		r.failed = true
		return 0, err
	}
	if shouldRefresh {
		return 1, nil
	}
	return 0, nil
}
