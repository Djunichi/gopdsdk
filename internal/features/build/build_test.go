package build

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRenderGoModForSameModule(t *testing.T) {
	sdkDir := filepath.Join(t.TempDir(), "Work tree", "gopdsdk")
	info := module{Path: sdkModule, Dir: sdkDir, GoVersion: "1.26.5"}
	goMod := renderGoMod(info, info)
	for _, want := range []string{
		"module github.com/Djunichi/gopdsdk/build",
		"go 1.26.5",
		"require github.com/Djunichi/gopdsdk v0.0.0",
		"replace github.com/Djunichi/gopdsdk => " + strconv.Quote(filepath.ToSlash(sdkDir)),
	} {
		if !strings.Contains(goMod, want) {
			t.Errorf("renderGoMod() does not contain %q:\n%s", want, goMod)
		}
	}
}

func TestRenderGoModForExternalApplication(t *testing.T) {
	sdkDir := filepath.Join(t.TempDir(), "SDK")
	gameDir := filepath.Join(t.TempDir(), "Game")
	goMod := renderGoMod(
		module{Path: sdkModule, Dir: sdkDir, GoVersion: "1.26.5"},
		module{Path: "example.com/game", Dir: gameDir, GoVersion: "1.26.5"},
	)
	for _, want := range []string{
		"require example.com/game v0.0.0",
		"replace example.com/game => " + strconv.Quote(filepath.ToSlash(gameDir)),
	} {
		if !strings.Contains(goMod, want) {
			t.Errorf("renderGoMod() does not contain %q:\n%s", want, goMod)
		}
	}
}
