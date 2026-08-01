package runtime

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/Djunichi/gopdsdk/playdate"
)

type testContext struct{}

func (testContext) Clear()                                                             {}
func (testContext) DrawText(string, int, int)                                          {}
func (testContext) CurrentTimeMilliseconds() uint32                                    { return 0 }
func (testContext) Input() playdate.Input                                              { return playdate.Input{} }
func (testContext) LoadBitmap(string) (playdate.Bitmap, error)                         { return nil, nil }
func (testContext) NewBitmap(int, int) (playdate.Bitmap, error)                        { return nil, nil }
func (testContext) DrawBitmap(playdate.Bitmap, int, int) error                         { return nil }
func (testContext) DrawScaledBitmap(playdate.Bitmap, int, int, float32, float32) error { return nil }

func TestBitmapOwnershipLifecycle(t *testing.T) {
	var freed []uintptr
	var fills []playdate.Color
	driver := BitmapDriver{
		Dimensions: func(uintptr) (int, int) { return 400, 240 },
		Fill:       func(_ uintptr, color playdate.Color) { fills = append(fills, color) },
		Free:       func(handle uintptr) { freed = append(freed, handle) },
	}
	owned := NewOwnedBitmap(7, driver)
	if width, err := owned.Width(); err != nil || width != 400 {
		t.Fatalf("Width() = %d, %v", width, err)
	}
	if height, err := owned.Height(); err != nil || height != 240 {
		t.Fatalf("Height() = %d, %v", height, err)
	}
	if err := owned.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := owned.Fill(playdate.ColorBlack); err != nil {
		t.Fatal(err)
	}
	if err := owned.Fill(playdate.Color(99)); !errors.Is(err, playdate.ErrBitmapColor) {
		t.Fatalf("invalid Fill() error = %v", err)
	}
	if err := owned.Close(); err != nil {
		t.Fatal(err)
	}
	if len(freed) != 1 || freed[0] != 7 {
		t.Fatalf("freed = %v", freed)
	}
	if len(fills) != 2 || fills[0] != playdate.ColorClear || fills[1] != playdate.ColorBlack {
		t.Fatalf("fills = %v", fills)
	}
	for name, operation := range map[string]func() error{
		"double close":           owned.Close,
		"fill after close":       func() error { return owned.Fill(playdate.ColorWhite) },
		"dimensions after close": func() error { _, err := owned.Width(); return err },
	} {
		if err := operation(); !errors.Is(err, playdate.ErrBitmapClosed) {
			t.Fatalf("%s error = %v", name, err)
		}
	}
	borrowed := NewBorrowedBitmap(8, driver)
	if err := borrowed.Close(); !errors.Is(err, playdate.ErrBitmapBorrowed) {
		t.Fatalf("borrowed Close() = %v", err)
	}
	if _, err := BitmapHandle(borrowed); err != nil {
		t.Fatal(err)
	}
}

func TestBitmapArgumentValidation(t *testing.T) {
	if err := ValidateBitmapSize(0, 1); !errors.Is(err, playdate.ErrBitmapSize) {
		t.Fatalf("size error = %v", err)
	}
	if err := ValidateBitmapScale(1, 0); !errors.Is(err, playdate.ErrBitmapScale) {
		t.Fatalf("scale error = %v", err)
	}
	if err := ValidateBitmapScale(1, *(*float32)(unsafe.Pointer(&[]uint32{0x7fc00000}[0]))); !errors.Is(err, playdate.ErrBitmapScale) {
		t.Fatalf("NaN scale error = %v", err)
	}
}

type testGame struct {
	init   func(playdate.Context) error
	update func(playdate.Context) (bool, error)
}

func (g testGame) Init(context playdate.Context) error { return g.init(context) }
func (g testGame) Update(context playdate.Context) (bool, error) {
	return g.update(context)
}

func TestABIFixedWidthTypes(t *testing.T) {
	if got := unsafe.Sizeof(Event(0)); got != 4 {
		t.Fatalf("sizeof(Event) = %d, want 4", got)
	}
}

func TestNewRequiresCallbacks(t *testing.T) {
	if _, err := New(Callbacks{}); !errors.Is(err, ErrInitRequired) {
		t.Fatalf("New() error = %v, want ErrInitRequired", err)
	}
	if _, err := New(Callbacks{Init: func() error { return nil }}); !errors.Is(err, ErrUpdateRequired) {
		t.Fatalf("New() error = %v, want ErrUpdateRequired", err)
	}
}

func TestRuntimeLifecycle(t *testing.T) {
	initCalls := 0
	updateCalls := 0
	runtime, err := New(Callbacks{
		Init: func() error {
			initCalls++
			return nil
		},
		Update: func(playdate.Input) (bool, error) {
			updateCalls++
			return true, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Update() before init error = %v, want ErrNotInitialized", err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	refresh, err := runtime.Update(RawInput{})
	if err != nil {
		t.Fatal(err)
	}
	if refresh != 1 || initCalls != 1 || updateCalls != 1 {
		t.Fatalf("refresh/init/update = %d/%d/%d, want 1/1/1", refresh, initCalls, updateCalls)
	}
	if err := runtime.Handle(EventInit, 0); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("duplicate Handle(EventInit) error = %v, want ErrAlreadyInitialized", err)
	}
}

func TestRuntimeRefreshAndErrors(t *testing.T) {
	updateErr := errors.New("update failed")
	tests := []struct {
		name        string
		update      func(playdate.Input) (bool, error)
		wantRefresh int32
		wantErr     error
	}{
		{"no refresh", func(playdate.Input) (bool, error) { return false, nil }, 0, nil},
		{"update error", func(playdate.Input) (bool, error) { return true, updateErr }, 0, updateErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime, err := New(Callbacks{Init: func() error { return nil }, Update: test.update})
			if err != nil {
				t.Fatal(err)
			}
			if err := runtime.Handle(EventInit, 0); err != nil {
				t.Fatal(err)
			}
			refresh, err := runtime.Update(RawInput{})
			if refresh != test.wantRefresh || !errors.Is(err, test.wantErr) {
				t.Fatalf("Update() = %d, %v; want %d, %v", refresh, err, test.wantRefresh, test.wantErr)
			}
		})
	}
}

func TestRuntimeFailedInitializationCannotUpdate(t *testing.T) {
	initErr := errors.New("init failed")
	runtime, err := New(Callbacks{
		Init:   func() error { return initErr },
		Update: func(playdate.Input) (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); !errors.Is(err, initErr) {
		t.Fatalf("Handle() error = %v, want init error", err)
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrFailed) {
		t.Fatalf("Update() error = %v, want ErrFailed", err)
	}
}

func TestApplicationIsCommonLifecycleEntryPoint(t *testing.T) {
	var calls []string
	context := testContext{}
	application, err := NewApplication(testGame{
		init: func(got playdate.Context) error {
			calls = append(calls, "init")
			return nil
		},
		update: func(got playdate.Context) (bool, error) {
			calls = append(calls, "update")
			return true, nil
		},
	}, context, func() { calls = append(calls, "before-init") })
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	refresh, err := application.Update(RawInput{})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := refresh, int32(1); got != want {
		t.Fatalf("Update() refresh = %d, want %d", got, want)
	}
	wantCalls := []string{"before-init", "init", "update"}
	if len(calls) != len(wantCalls) {
		t.Fatalf("calls = %v, want %v", calls, wantCalls)
	}
	for index := range wantCalls {
		if calls[index] != wantCalls[index] {
			t.Fatalf("calls = %v, want %v", calls, wantCalls)
		}
	}
}

func TestLifecycleEventsPreservePlatformOrder(t *testing.T) {
	var got []playdate.LifecycleEvent
	runtime, err := New(Callbacks{
		Init: func() error { return nil },
		Lifecycle: func(event playdate.LifecycleEvent) error {
			got = append(got, event)
			return nil
		},
		Update: func(playdate.Input) (bool, error) { return false, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	events := []Event{EventPause, EventLock, EventLowPower, EventUnlock, EventResume, EventTerminate}
	for _, event := range events {
		if err := runtime.Handle(event, 0); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrTerminated) {
		t.Fatalf("Update() after terminate error = %v, want ErrTerminated", err)
	}
	want := []playdate.LifecycleEvent{playdate.LifecyclePause, playdate.LifecycleLock, playdate.LifecycleLowPower, playdate.LifecycleUnlock, playdate.LifecycleResume, playdate.LifecycleTerminate}
	if len(got) != len(want) {
		t.Fatalf("events = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("events = %v, want %v", got, want)
		}
	}
}

func TestLifecycleErrorIsFailStop(t *testing.T) {
	wantErr := errors.New("pause failed")
	runtime, err := New(Callbacks{
		Init:      func() error { return nil },
		Lifecycle: func(playdate.LifecycleEvent) error { return wantErr },
		Update:    func(playdate.Input) (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventPause, 0); !errors.Is(err, wantErr) {
		t.Fatalf("pause error = %v, want %v", err, wantErr)
	}
	if _, err := runtime.Update(RawInput{}); !errors.Is(err, ErrFailed) {
		t.Fatalf("Update() error = %v, want ErrFailed", err)
	}
}

func TestInputTransitionsBetweenFrames(t *testing.T) {
	var snapshots []playdate.Input
	runtime, err := New(Callbacks{
		Init: func() error { return nil },
		Update: func(input playdate.Input) (bool, error) {
			snapshots = append(snapshots, input)
			return false, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	frames := []RawInput{
		{Buttons: playdate.ButtonA, CrankAngle: 10, CrankDelta: 2, CrankDocked: true, DeltaSeconds: 0.016},
		{Buttons: playdate.ButtonA | playdate.ButtonLeft, CrankAngle: 15, CrankDelta: 5, CrankDocked: false, DeltaSeconds: 0.017},
		{Buttons: playdate.ButtonLeft, CrankAngle: 15, CrankDocked: true, DeltaSeconds: 0.018},
	}
	for _, frame := range frames {
		if _, err := runtime.Update(frame); err != nil {
			t.Fatal(err)
		}
	}
	if got := snapshots[0]; got.Pressed != playdate.ButtonA || got.Held != 0 || got.Released != 0 || got.CrankDockedThisFrame || got.CrankUndocked {
		t.Fatalf("first snapshot transitions = %+v", got)
	}
	if got := snapshots[1]; got.Pressed != playdate.ButtonLeft || got.Held != playdate.ButtonA || got.Released != 0 || !got.CrankUndocked || got.DeltaSeconds != 0.017 {
		t.Fatalf("second snapshot transitions = %+v", got)
	}
	if got := snapshots[2]; got.Pressed != 0 || got.Held != playdate.ButtonLeft || got.Released != playdate.ButtonA || !got.CrankDockedThisFrame {
		t.Fatalf("third snapshot transitions = %+v", got)
	}
}
