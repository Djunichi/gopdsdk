package doctor

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestWriteReport(t *testing.T) {
	report := Report{
		Host:       "windows/amd64",
		SDKPath:    `C:\Users\dev\Documents\PlaydateSDK`,
		SDKVersion: "3.1.1",
		Capabilities: []Capability{
			{"sdk", StatusReady, "Playdate SDK 3.1.1"},
			{"device-build", StatusMissing, "TinyGo not found"},
		},
	}
	var output bytes.Buffer
	if err := writeReport(&output, report); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"Host:         windows/amd64",
		"Playdate SDK: 3.1.1",
		`SDK path:     C:\Users\dev\Documents\PlaydateSDK`,
		"sdk            READY",
		"device-build   MISSING",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("output does not contain %q:\n%s", want, output.String())
		}
	}
}

func TestRunRejectsUnknownCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := Run(t.Context(), []string{"build"}, &stdout, &stderr, Options{})
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("Run error = %v, want unknown command", err)
	}
}

func TestRunSimulatorProbeReady(t *testing.T) {
	report := Report{
		SDKPath: "sdk",
		Capabilities: []Capability{
			{Name: "simulator", Status: StatusUnverified, Summary: "gcc found"},
		},
	}
	called := false
	runSimulatorProbe(t.Context(), &report, func(_ context.Context, sdkPath string) error {
		called = true
		if sdkPath != "sdk" {
			t.Fatalf("sdkPath = %q, want sdk", sdkPath)
		}
		return nil
	})
	if !called {
		t.Fatal("simulator probe was not called")
	}
	if got := report.Capabilities[0].Status; got != StatusReady {
		t.Fatalf("simulator status = %q, want ready", got)
	}
}

func TestRunDeviceProbeReady(t *testing.T) {
	report := Report{
		SDKPath: "sdk",
		Capabilities: []Capability{
			{Name: "device-build", Status: StatusUnverified, Summary: "toolchain found"},
		},
	}
	called := false
	runDeviceProbe(t.Context(), &report, func(_ context.Context, sdkPath string) error {
		called = true
		if sdkPath != "sdk" {
			t.Fatalf("sdkPath = %q, want sdk", sdkPath)
		}
		return nil
	})
	if !called {
		t.Fatal("device probe was not called")
	}
	if got := report.Capabilities[0].Status; got != StatusReady {
		t.Fatalf("device-build status = %q, want ready", got)
	}
	if !strings.Contains(report.Capabilities[0].Summary, "TinyGo PIC build") {
		t.Fatalf("device-build summary = %q", report.Capabilities[0].Summary)
	}
}

func TestRunCapabilityProbeFailureDoesNotChangeOtherCapabilities(t *testing.T) {
	report := Report{Capabilities: []Capability{
		{Name: "simulator", Status: StatusUnverified, Summary: "gcc found"},
		{Name: "device-build", Status: StatusUnverified, Summary: "toolchain found"},
	}}
	runSimulatorProbe(t.Context(), &report, func(context.Context, string) error {
		return fmt.Errorf("probe failed")
	})
	if got := report.Capabilities[0].Status; got != StatusIncompatible {
		t.Fatalf("simulator status = %q, want incompatible", got)
	}
	if got := report.Capabilities[1].Status; got != StatusUnverified {
		t.Fatalf("device-build status = %q, want unverified", got)
	}
}
