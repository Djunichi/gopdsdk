package build

import (
	"strings"
	"testing"
)

func TestRenderGoModForSameModule(t *testing.T) {
	info := module{Path: sdkModule, Dir: `C:\Work tree\gopdsdk`, GoVersion: "1.26.5"}
	goMod := renderGoMod(info, info)
	for _, want := range []string{
		"module github.com/Djunichi/gopdsdk/build",
		"go 1.26.5",
		"require github.com/Djunichi/gopdsdk v0.0.0",
		`replace github.com/Djunichi/gopdsdk => "C:/Work tree/gopdsdk"`,
	} {
		if !strings.Contains(goMod, want) {
			t.Errorf("renderGoMod() does not contain %q:\n%s", want, goMod)
		}
	}
}

func TestRenderGoModForExternalApplication(t *testing.T) {
	goMod := renderGoMod(
		module{Path: sdkModule, Dir: `C:\SDK`, GoVersion: "1.26.5"},
		module{Path: "example.com/game", Dir: `C:\Game`, GoVersion: "1.26.5"},
	)
	for _, want := range []string{
		"require example.com/game v0.0.0",
		`replace example.com/game => "C:/Game"`,
	} {
		if !strings.Contains(goMod, want) {
			t.Errorf("renderGoMod() does not contain %q:\n%s", want, goMod)
		}
	}
}

func TestRenderPDXInfo(t *testing.T) {
	info := renderPDXInfo("hello")
	if !strings.Contains(info, "name=hello\n") || !strings.Contains(info, "bundleID=sdk.gopdsdk.hello\n") {
		t.Fatalf("renderPDXInfo() = %q", info)
	}
}
