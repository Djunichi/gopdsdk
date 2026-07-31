package devicecrashlog

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestIsPlaydateRoot(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"Data", "Games", "System"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "crashlog.txt"), []byte("crash"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !isPlaydateRoot(root) {
		t.Fatal("isPlaydateRoot() = false, want true")
	}
}

func TestIsPlaydateRootRejectsUnrelatedVolume(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "crashlog.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if isPlaydateRoot(root) {
		t.Fatal("isPlaydateRoot() = true for unrelated volume")
	}
}

func TestParseMountPath(t *testing.T) {
	path := "/Volumes/PLAYDATE"
	if runtime.GOOS == "windows" {
		path = `F:\`
	}
	output := "Playdate device detected\r\nWaiting for drive to appear...\r\nPlaydate data disk mounted as " + path + "\r\n"
	got, err := parseMountPath(output)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Clean(path); got != want {
		t.Fatalf("parseMountPath() = %q, want %q", got, want)
	}
}

func TestParseMountPathRejectsMissingPath(t *testing.T) {
	_, err := parseMountPath("Playdate device detected\n")
	if err == nil || !strings.Contains(err.Error(), "did not report") {
		t.Fatalf("parseMountPath() error = %v, want missing-path error", err)
	}
}

func TestParseMountPathRejectsRelativePath(t *testing.T) {
	_, err := parseMountPath("Playdate data disk mounted as relative/path\n")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("parseMountPath() error = %v, want invalid-path error", err)
	}
}
