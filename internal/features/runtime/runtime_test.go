package runtime

import (
	"errors"
	"testing"
	"unsafe"
)

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
		Update: func() (bool, error) {
			updateCalls++
			return true, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := runtime.Update(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Update() before init error = %v, want ErrNotInitialized", err)
	}
	if err := runtime.Handle(EventInit, 0); err != nil {
		t.Fatal(err)
	}
	refresh, err := runtime.Update()
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
		update      func() (bool, error)
		wantRefresh int32
		wantErr     error
	}{
		{"no refresh", func() (bool, error) { return false, nil }, 0, nil},
		{"update error", func() (bool, error) { return true, updateErr }, 0, updateErr},
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
			refresh, err := runtime.Update()
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
		Update: func() (bool, error) { return true, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.Handle(EventInit, 0); !errors.Is(err, initErr) {
		t.Fatalf("Handle() error = %v, want init error", err)
	}
	if _, err := runtime.Update(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Update() error = %v, want ErrNotInitialized", err)
	}
}
