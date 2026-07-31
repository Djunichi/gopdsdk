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

func TestTargetUsesHardFloatCortexM7(t *testing.T) {
	for _, want := range []string{"cortex-m7", "thumbv7em-unknown-unknown-eabihf", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16"} {
		if !strings.Contains(targetSource, want) {
			t.Errorf("targetSource does not contain %q", want)
		}
	}
}

func TestAdapterDefinesInterruptHooks(t *testing.T) {
	for _, want := range []string{"DisableInterrupts:", "mrs r0, primask", "EnableInterrupts:", "msr primask, r0"} {
		if !strings.Contains(adapterSource, want) {
			t.Errorf("adapterSource does not contain %q", want)
		}
	}
}

func TestFirstLine(t *testing.T) {
	if got, want := firstLine("first\r\nsecond\n"), "first"; got != want {
		t.Fatalf("firstLine() = %q, want %q", got, want)
	}
}
