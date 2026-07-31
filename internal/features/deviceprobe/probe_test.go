package deviceprobe

import (
	"strings"
	"testing"
)

func TestProbeSourceExportsEventHandler(t *testing.T) {
	for _, want := range []string{"package main", "//export eventHandler", "func eventHandler", "func main()"} {
		if !strings.Contains(probeSource, want) {
			t.Errorf("probeSource does not contain %q", want)
		}
	}
}

func TestFirstLine(t *testing.T) {
	if got, want := firstLine("first\r\nsecond\n"), "first"; got != want {
		t.Fatalf("firstLine() = %q, want %q", got, want)
	}
}
