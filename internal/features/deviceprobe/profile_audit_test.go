package deviceprobe

import (
	"errors"
	"strings"
	"testing"
)

func TestProfileProbeCatalogCoversP12AuditSurface(t *testing.T) {
	var names []string
	for _, probe := range profileProbes() {
		names = append(names, probe.name)
		if !strings.HasPrefix(probe.source, "package main\n") {
			t.Errorf("probe %q is not a main package", probe.name)
		}
	}
	for _, want := range []string{"baseline", "goroutine", "channel", "select", "defer", "panic", "recover", "reflection", "finalizer", "cgo", "runtime-gc", "stdlib-time", "stdlib-fmt", "stdlib-json"} {
		if !containsProfileProbe(names, want) {
			t.Errorf("profile probe catalog is missing %q", want)
		}
	}
}

func TestClassifyProfileProbe(t *testing.T) {
	tests := []struct {
		name       string
		compileErr error
		output     string
		symbols    string
		status     ProfileStatus
	}{
		{name: "build", status: ProfileBuildOnly},
		{name: "rejected", compileErr: errors.New("exit"), output: "compile failed", status: ProfileRejected},
		{name: "unsafe", symbols: "00001000 t runtime.chanSend\n", status: ProfileUnsafe},
	}
	for _, test := range tests {
		result := classifyProfileProbe(test.name, test.output, test.compileErr, test.symbols)
		if result.Name != test.name || result.Status != test.status || result.Evidence == "" {
			t.Errorf("classifyProfileProbe(%q) = %+v, want status %s", test.name, result, test.status)
		}
	}
}

func TestBaselineIsFirstProfileProbe(t *testing.T) {
	probes := profileProbes()
	if len(probes) == 0 || probes[0].name != "baseline" {
		t.Fatalf("first profile probe = %v, want baseline", probes)
	}
}

func TestConciseEvidenceIsBounded(t *testing.T) {
	if got := conciseEvidence(""); got == "" {
		t.Fatal("empty tool output produced empty evidence")
	}
	if got := conciseEvidence(strings.Repeat("word ", 100)); len(got) > 240 {
		t.Fatalf("conciseEvidence length = %d", len(got))
	}
}

func containsProfileProbe(names []string, want string) bool {
	for _, name := range names {
		if name == want {
			return true
		}
	}
	return false
}
