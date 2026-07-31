package buildplan

import (
	"bytes"
	"strings"
	"testing"
)

func TestDevicePlanIsDeterministicAndStructured(t *testing.T) {
	plan, err := New(Device, "./game", `C:\Playdate SDK`, `build\game.pdx`)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Commands) != 9 {
		t.Fatalf("len(Commands) = %d, want 9", len(plan.Commands))
	}
	if plan.Commands[0].Executable != "tinygo" || plan.Commands[0].Args[0] != "build" {
		t.Fatalf("first command = %#v", plan.Commands[0])
	}
	var output bytes.Buffer
	if err := Write(&output, plan); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Target:      device", `Application: ./game`, `"tinygo" "build"`, `"arm-none-eabi-readelf"`, `${WORK}/pdex.elf`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("plan does not contain %q:\n%s", want, output.String())
		}
	}
}

func TestSimulatorPlanDoesNotRequireSDKExecution(t *testing.T) {
	plan, err := New(Simulator, ".", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := plan.SDKPath, "${PLAYDATE_SDK_PATH}"; got != want {
		t.Fatalf("SDKPath = %q, want %q", got, want)
	}
	if len(plan.Commands) != 2 || plan.Commands[0].Executable != "go" {
		t.Fatalf("Commands = %#v", plan.Commands)
	}
}

func TestNewRejectsUnknownTarget(t *testing.T) {
	if _, err := New(Target("unknown"), ".", "", ""); err == nil {
		t.Fatal("New() error = nil, want unknown-target error")
	}
}
