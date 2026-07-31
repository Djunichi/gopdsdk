package doctor

import "testing"

func TestMatchesVerifiedTool(t *testing.T) {
	tests := []struct{ name, version string }{
		{"go", "go version go1.26.5 windows/amd64"},
		{"tinygo", "tinygo version 0.41.1 windows/amd64"},
		{"arm-none-eabi-gcc", "arm-none-eabi-gcc.exe (Arm GNU Toolchain) 15.3.1 20260627"},
	}
	for _, test := range tests {
		if !matchesVerifiedTool(test.name, test.version) {
			t.Errorf("matchesVerifiedTool(%q, %q) = false", test.name, test.version)
		}
	}
	if matchesVerifiedTool("go", "go version go1.27.0 darwin/arm64") {
		t.Fatal("unverified Go version matched")
	}
}
