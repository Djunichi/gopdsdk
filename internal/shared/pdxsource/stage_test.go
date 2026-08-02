package pdxsource

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStageCopiesOnlyResourceContentsToPDXRoot(t *testing.T) {
	application := t.TempDir()
	source := t.TempDir()
	files := map[string]string{
		"game.go": "package game", "game_test.go": "package game", "go.mod": "module game",
		"pdxinfo": "name=game\nimagePath=images/launcher", "README.md": "private notes",
		"resources/images/icon.png": "png", "resources/audio/theme.wav": "wav", "resources/data/level.json": "{}",
		"resources/images/launcher/card.png": "card", "resources/images/launcher/icon.png": "icon", "resources/images/launcher/launchImage.png": "launch",
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
	for sourcePath, resourcePath := range map[string]string{
		"images/icon.png": "resources/images/icon.png", "audio/theme.wav": "resources/audio/theme.wav", "data/level.json": "resources/data/level.json",
		"images/launcher/card.png": "resources/images/launcher/card.png", "images/launcher/icon.png": "resources/images/launcher/icon.png", "images/launcher/launchImage.png": "resources/images/launcher/launchImage.png",
	} {
		if contents, err := os.ReadFile(filepath.Join(source, sourcePath)); err != nil || string(contents) != files[resourcePath] {
			t.Fatalf("resource %s = %q, %v", sourcePath, contents, err)
		}
	}
	for _, path := range []string{"game.go", "game_test.go", "go.mod", "pdxinfo", "README.md", "resources"} {
		if _, err := os.Stat(filepath.Join(source, path)); !os.IsNotExist(err) {
			t.Fatalf("excluded %s exists or stat failed: %v", path, err)
		}
	}
}

func TestStageAllowsMissingResourcesDirectory(t *testing.T) {
	if err := Stage(t.TempDir(), t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
