package gomodule

import (
	"strings"
	"testing"
)

func TestRenderProbe(t *testing.T) {
	got := RenderProbe(Info{Root: `C:\Work tree\gopdsdk`, Path: "github.com/Djunichi/gopdsdk", GoVersion: "1.26.5"}, "probe/device")
	for _, want := range []string{"module github.com/Djunichi/gopdsdk/probe/device", "go 1.26.5", "require github.com/Djunichi/gopdsdk v0.0.0", `replace github.com/Djunichi/gopdsdk => "C:/Work tree/gopdsdk"`} {
		if !strings.Contains(got, want) {
			t.Errorf("RenderProbe() does not contain %q:\n%s", want, got)
		}
	}
}
