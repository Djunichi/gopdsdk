package devicelog

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReadRejectsUnsupportedLog(t *testing.T) {
	_, _, err := Read(context.Background(), "sdk", Kind("other.txt"))
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("Read(other.txt) error = %v, want unsupported-log error", err)
	}
}

func TestIsPlaydateRootForRequestedLog(t *testing.T) {
	root := makePlaydateRoot(t)
	if err := os.WriteFile(filepath.Join(root, string(ErrorLog)), []byte("error"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !isPlaydateRoot(root) {
		t.Fatal("isPlaydateRoot() = false, want true")
	}
	if _, _, err := readLog(root, CrashLog); err == nil || !strings.Contains(err.Error(), "crashlog.txt") {
		t.Fatalf("readLog(crashlog) error = %v, want missing crashlog.txt", err)
	}
}

func TestIsPlaydateRootRejectsUnrelatedVolume(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, string(CrashLog)), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if isPlaydateRoot(root) {
		t.Fatal("isPlaydateRoot() = true for unrelated volume")
	}
}

func TestReadLog(t *testing.T) {
	root := t.TempDir()
	want := []byte("diagnostic contents")
	if err := os.WriteFile(filepath.Join(root, string(ErrorLog)), want, 0o600); err != nil {
		t.Fatal(err)
	}
	got, path, err := readLog(root, ErrorLog)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) || path != filepath.Join(root, string(ErrorLog)) {
		t.Fatalf("readLog() = %q, %q; want %q, requested path", got, path, want)
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

func TestCommandKind(t *testing.T) {
	for command, want := range map[string]Kind{"crashlog": CrashLog, "errorlog": ErrorLog} {
		got, ok := commandKind(command)
		if !ok || got != want {
			t.Errorf("commandKind(%q) = %q, %v; want %q, true", command, got, ok, want)
		}
	}
	if _, ok := commandKind("other"); ok {
		t.Fatal("commandKind(other) = true")
	}
}

func makePlaydateRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{"Data", "Games", "System"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	return root
}
