package initproject

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCreateWritesBuildableProjectContract(t *testing.T) {
	path := filepath.Join(t.TempDir(), "My Game")
	sdkDir := filepath.Join(t.TempDir(), "Work tree", "gopdsdk")
	result, err := Create(Config{Path: path, SDKDir: sdkDir, GoVersion: "1.26.5"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Module != "example.com/my-game" {
		t.Fatalf("module = %q", result.Module)
	}
	wants := map[string][]string{
		"go.mod":  {"module example.com/my-game", "replace github.com/Djunichi/gopdsdk => " + strconv.Quote(filepath.ToSlash(sdkDir))},
		"game.go": {"// Package game", "func New() playdate.Game", "context.DrawText"},
		"pdxinfo": {"name=My Game", "author=Your Name", "bundleID=com.example.my-game", "buildNumber=1"},
	}
	for name, fragments := range wants {
		contents, readErr := os.ReadFile(filepath.Join(path, name))
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(contents), fragment) {
				t.Errorf("%s does not contain %q:\n%s", name, fragment, contents)
			}
		}
	}
}

func TestCreateDoesNotModifyExistingPath(t *testing.T) {
	path := t.TempDir()
	marker := filepath.Join(path, "keep")
	if err := os.WriteFile(marker, []byte("user data"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Create(Config{Path: path, SDKDir: "sdk", GoVersion: "1.26.5"})
	if err == nil {
		t.Fatal("Create() error = nil")
	}
	if contents, readErr := os.ReadFile(marker); readErr != nil || string(contents) != "user data" {
		t.Fatalf("existing data changed: %q, %v", contents, readErr)
	}
}

func TestProjectSlug(t *testing.T) {
	if got, want := projectSlug(" My  First_Game! "), "my-first-game"; got != want {
		t.Fatalf("projectSlug() = %q, want %q", got, want)
	}
}
