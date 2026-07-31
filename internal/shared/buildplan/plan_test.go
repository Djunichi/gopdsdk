package buildplan

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevicePlanIsDeterministicAndStructured(t *testing.T) {
	plan, err := New(Device, "./game", `C:\Playdate SDK`, `build\game.pdx`)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Commands) != 11 {
		t.Fatalf("len(Commands) = %d, want 11", len(plan.Commands))
	}
	if plan.Commands[1].Executable != "tinygo" || plan.Commands[1].Args[0] != "build" {
		t.Fatalf("TinyGo command = %#v", plan.Commands[1])
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

func TestResolveSubstitutesStructuredCommandFields(t *testing.T) {
	plan, err := New(Simulator, ".", "sdk", "out")
	if err != nil {
		t.Fatal(err)
	}
	resolved := Resolve(plan, map[string]string{"${WORK}": "work", "${HOST_LIBRARY}": "dll", "${CC}": "gcc", "${PDC}": "pdc.exe", "${PACKAGE_OUTPUT}": "temporary.pdx"})
	if got, want := resolved.Commands[0].Args[3], "work/Source/pdex.dll"; got != want {
		t.Fatalf("library output = %q, want %q", got, want)
	}
	if got, want := resolved.Commands[0].Environment[1], "CC=gcc"; got != want {
		t.Fatalf("compiler environment = %q, want %q", got, want)
	}
	if got, want := resolved.Commands[1].Executable, "pdc.exe"; got != want {
		t.Fatalf("pdc executable = %q, want %q", got, want)
	}
}

func TestBindExecutablesPreservesArguments(t *testing.T) {
	plan, err := New(Device, ".", "sdk", "out")
	if err != nil {
		t.Fatal(err)
	}
	bound := BindExecutables(plan, map[string]string{"tinygo": `C:\tools\tinygo.exe`})
	if got, want := bound.Commands[1].Executable, `C:\tools\tinygo.exe`; got != want {
		t.Fatalf("Executable = %q, want %q", got, want)
	}
	if got, want := bound.Commands[1].Args[0], "build"; got != want {
		t.Fatalf("first argument = %q, want %q", got, want)
	}
	if plan.Commands[1].Executable != "tinygo" {
		t.Fatal("BindExecutables mutated source plan")
	}
}

func TestCleanupPathsReturnsOnlyResolvedOwnedArtifacts(t *testing.T) {
	published := filepath.Join(t.TempDir(), "published", "game.pdx")
	workspace := filepath.Join(t.TempDir(), "gopdsdk-work")
	plan, err := New(Device, ".", "sdk", published)
	if err != nil {
		t.Fatal(err)
	}
	plan = Resolve(plan, map[string]string{"${WORK}": workspace})
	paths, err := CleanupPaths(plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != workspace {
		t.Fatalf("CleanupPaths() = %v", paths)
	}
}

func TestCleanupPathsRejectsUnresolvedAndRootPaths(t *testing.T) {
	for _, path := range []string{"${WORK}", string(filepath.Separator)} {
		plan := Plan{Artifacts: []Artifact{{Name: "workspace", Path: path, Retention: Cleanup}}}
		if _, err := CleanupPaths(plan); err == nil {
			t.Fatalf("CleanupPaths(%q) error = nil", path)
		}
	}
}
