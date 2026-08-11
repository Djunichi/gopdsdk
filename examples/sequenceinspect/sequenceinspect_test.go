package sequenceinspect

import "testing"

func TestFormatting(t *testing.T) {
	if small(12) != "12" || signed(-3) != "-3" || yes(true) != "yes" {
		t.Fatal("formatting")
	}
}
