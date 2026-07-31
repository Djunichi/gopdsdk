package deviceconnect

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyNoDevice(t *testing.T) {
	_, err := classify("pdutil", "No Playdate device detected.\r\n", errors.New("exit status 1"))
	if !errors.Is(err, ErrNoDevice) {
		t.Fatalf("classify() error = %v, want ErrNoDevice", err)
	}
	if !strings.Contains(err.Error(), "connect and unlock") {
		t.Fatalf("classify() error = %q, want remediation", err)
	}
}

func TestClassifyConnected(t *testing.T) {
	result, err := classify("pdutil", "Usage: pdutil <action> [options]\n", nil)
	if err != nil {
		t.Fatalf("classify() error = %v", err)
	}
	if result.Tool != "pdutil" || !strings.Contains(result.Status, "without performing an action") {
		t.Fatalf("classify() = %+v", result)
	}
}

func TestClassifyConnectedBeforeUnknownAction(t *testing.T) {
	result, err := classify("pdutil", "Playdate device detected on COM3\r\nUnknown action \"--help\".\r\n", errors.New("exit status 2"))
	if err != nil {
		t.Fatalf("classify() error = %v", err)
	}
	if !strings.Contains(result.Status, "COM3") || !strings.Contains(result.Status, "no device action") {
		t.Fatalf("classify() = %+v", result)
	}
}

func TestClassifyRejectsUnknownFailure(t *testing.T) {
	_, err := classify("pdutil", "access denied", errors.New("exit status 1"))
	if err == nil || errors.Is(err, ErrNoDevice) {
		t.Fatalf("classify() error = %v, want non-device failure", err)
	}
}

func TestClassifyRejectsUnknownSuccess(t *testing.T) {
	_, err := classify("pdutil", "unexpected", nil)
	if err == nil {
		t.Fatal("classify() error = nil, want unrecognized response")
	}
}
