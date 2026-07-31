package hostpolicy

import (
	"strings"
	"testing"
)

func TestPolicies(t *testing.T) {
	tests := []struct{ goos, library, pdc, simulator string }{
		{"windows", "dll", "pdc.exe", "PlaydateSimulator.exe"},
		{"darwin", "dylib", "pdc", "Playdate Simulator"},
		{"linux", "so", "pdc", "PlaydateSimulator"},
	}
	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			policy, err := For(test.goos)
			if err != nil {
				t.Fatal(err)
			}
			if policy.LibraryExtension != test.library || policy.PDCName != test.pdc || !strings.Contains(policy.SimulatorCandidates[0], test.simulator) {
				t.Fatalf("Policy = %#v", policy)
			}
		})
	}
}

func TestForRejectsUnsupportedHost(t *testing.T) {
	if _, err := For("plan9"); err == nil {
		t.Fatal("For() error = nil")
	}
}
