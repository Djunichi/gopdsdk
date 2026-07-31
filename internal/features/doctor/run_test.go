package doctor

import (
	"bytes"
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
	err := Run(t.Context(), []string{"build"}, &stdout, &stderr)
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("Run error = %v, want unknown command", err)
	}
}
