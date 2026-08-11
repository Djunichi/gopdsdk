package synthesis

import "testing"

func TestCurveName(t *testing.T) {
	for value, want := range map[float32]string{-1: "-1.0", -.3: "-0.3", 0: "0.0 (linear)", .4: "+0.4", 1: "+1.0"} {
		if got := curveName(value); got != want {
			t.Errorf("curveName(%v) = %q, want %q", value, got, want)
		}
	}
}
