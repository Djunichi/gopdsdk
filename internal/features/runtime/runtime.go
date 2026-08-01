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
)

// Callbacks contains application behavior invoked by Runtime.
type Callbacks struct {
	Init   func() error
	Update func() (refresh bool, err error)
}

// Runtime enforces the platform-independent application lifecycle.
type Runtime struct {
	callbacks   Callbacks
	initialized bool
	failed      bool
}

// Application is the platform-independent entry point shared by Simulator and
// device ABI adapters.
type Application struct {
	runtime *Runtime
}

// NewApplication composes a public game with its platform context. beforeInit
// runs immediately before the game's initialization callback when a platform
// adapter needs to prepare callback state.
func NewApplication(game playdate.Game, context playdate.Context, beforeInit func()) (*Application, error) {
	runtime, err := New(Callbacks{
		Init: func() error {
			if beforeInit != nil {
				beforeInit()
			}
			return game.Init(context)
		},
		Update: func() (bool, error) {
			return game.Update(context)
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
func (a *Application) Update() (int32, error) {
	return a.runtime.Update()
}

// New validates callbacks and returns an uninitialized runtime.
func New(callbacks Callbacks) (*Runtime, error) {
	if callbacks.Init == nil {
		return nil, ErrInitRequired
	}
	if callbacks.Update == nil {
		return nil, ErrUpdateRequired
	}
	return &Runtime{callbacks: callbacks}, nil
}

// Handle delivers a Playdate system event to the lifecycle.
func (r *Runtime) Handle(event Event, _ uint32) error {
	if event != EventInit {
		return nil
	}
	if r.initialized || r.failed {
		return ErrAlreadyInitialized
	}
	if err := r.callbacks.Init(); err != nil {
		r.failed = true
		return err
	}
	r.initialized = true
	return nil
}

// Update invokes application update and converts its refresh decision to the
// integer contract expected by Playdate.
func (r *Runtime) Update() (int32, error) {
	if !r.initialized || r.failed {
		return 0, ErrNotInitialized
	}
	shouldRefresh, err := r.callbacks.Update()
	if err != nil {
		r.failed = true
		return 0, err
	}
	if shouldRefresh {
		return 1, nil
	}
	return 0, nil
}
