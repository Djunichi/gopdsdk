package pdxsource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStageCopiesResourcesAndExcludesBuildInputs(t *testing.T) {
	application := t.TempDir()
	source := t.TempDir()
	files := map[string]string{
		"game.go": "package game", "game_test.go": "package game", "go.mod": "module game",
		"pdxinfo": "name=game", "images/icon.png": "png", "data/level.json": "{}",
	}
	for path, contents := range files {
		fullPath := filepath.Join(application, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := Stage(application, source); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"images/icon.png", "data/level.json"} {
		if contents, err := os.ReadFile(filepath.Join(source, path)); err != nil || string(contents) != files[path] {
			t.Fatalf("resource %s = %q, %v", path, contents, err)
		}
	}
	for _, path := range []string{"game.go", "game_test.go", "go.mod", "pdxinfo"} {
		if _, err := os.Stat(filepath.Join(source, path)); !os.IsNotExist(err) {
			t.Fatalf("excluded %s exists or stat failed: %v", path, err)
		}
	}
}
