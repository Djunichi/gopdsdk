// Package runtime coordinates platform-independent Playdate application lifecycle.
package runtime

import (
	"errors"

	"github.com/Djunichi/gopdsdk/playdate"
)

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
	owned  bool
	closed bool
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
	if b == nil || b.closed || b.handle == 0 {
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
	input playdate.Input
}

func (context *applicationContext) Input() playdate.Input { return context.input }

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
			if lifecycle == nil {
				return nil
			}
			return lifecycle.HandleLifecycle(gameContext, event)
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
