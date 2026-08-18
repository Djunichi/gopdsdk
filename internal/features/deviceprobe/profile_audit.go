package deviceprobe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Djunichi/gopdsdk/internal/shared/buildplan"
)

// ProfileStatus is the strongest evidence produced by the compile-stage audit.
type ProfileStatus string

const (
	ProfileBuildOnly ProfileStatus = "build-only"
	ProfileRejected  ProfileStatus = "rejected"
	ProfileUnsafe    ProfileStatus = "unsafe"
)

// ProfileResult records one isolated device Go-profile probe.
type ProfileResult struct {
	Name     string        `json:"name"`
	Status   ProfileStatus `json:"status"`
	Evidence string        `json:"evidence"`
}

type profileProbe struct {
	name   string
	source string
}

// AuditProfile compiles isolated language and standard-library probes for the
// accepted TinyGo target. It does not claim Simulator or physical-device
// execution evidence.
func AuditProfile(ctx context.Context) ([]ProfileResult, error) {
	tinygo, err := exec.LookPath("tinygo")
	if err != nil {
		return nil, fmt.Errorf("find tinygo: %w", err)
	}
	nm, err := exec.LookPath("arm-none-eabi-nm")
	if err != nil {
		return nil, fmt.Errorf("find arm-none-eabi-nm: %w", err)
	}
	root, err := os.MkdirTemp("", "gopdsdk-device-profile-")
	if err != nil {
		return nil, fmt.Errorf("create device profile directory: %w", err)
	}
	defer os.RemoveAll(root)
	results := make([]ProfileResult, 0, len(profileProbes()))
	for _, probe := range profileProbes() {
		result, auditErr := auditProfileProbe(ctx, root, tinygo, nm, probe)
		if auditErr != nil {
			return nil, auditErr
		}
		if probe.name == "baseline" && result.Status != ProfileBuildOnly {
			return nil, fmt.Errorf("baseline device profile probe did not compile: %s", result.Evidence)
		}
		results = append(results, result)
	}
	return results, nil
}

func auditProfileProbe(ctx context.Context, root, tinygo, nm string, probe profileProbe) (ProfileResult, error) {
	directory := filepath.Join(root, probe.name)
	if err := os.Mkdir(directory, 0o755); err != nil {
		return ProfileResult{}, fmt.Errorf("create %s probe directory: %w", probe.name, err)
	}
	for name, contents := range map[string]string{
		"go.mod":        "module example.com/gopdsdk-device-profile/" + probe.name + "\n\ngo 1.26\n",
		"main.go":       probe.source,
		"playdate.json": targetSource,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			return ProfileResult{}, fmt.Errorf("write %s probe %s: %w", probe.name, name, err)
		}
	}
	object := filepath.Join(directory, "probe.o")
	compileOutput, compileErr := execProfileCommand(ctx, directory, tinygo, "build", "-target", filepath.Join(directory, "playdate.json"), "-scheduler", "none", "-gc", "conservative", "-panic", "trap", "-opt", "0", "-o", object, ".")
	if compileErr != nil {
		return classifyProfileProbe(probe.name, compileOutput, compileErr, ""), nil
	}
	symbolOutput, symbolErr := execProfileCommand(ctx, directory, nm, object)
	if symbolErr != nil {
		return ProfileResult{}, fmt.Errorf("inspect %s probe symbols: %w: %s", probe.name, symbolErr, conciseEvidence(symbolOutput))
	}
	return classifyProfileProbe(probe.name, compileOutput, nil, symbolOutput), nil
}

func execProfileCommand(ctx context.Context, directory, executable string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	return string(output), err
}

func classifyProfileProbe(name, compileOutput string, compileErr error, symbols string) ProfileResult {
	if compileErr != nil {
		return ProfileResult{Name: name, Status: ProfileRejected, Evidence: conciseEvidence(compileOutput)}
	}
	if unsupported := unsupportedRuntimeSymbols(symbols, buildplan.DeviceMemoryConservative); len(unsupported) != 0 {
		return ProfileResult{Name: name, Status: ProfileUnsafe, Evidence: "forbidden runtime symbols: " + strings.Join(unsupported, ", ")}
	}
	return ProfileResult{Name: name, Status: ProfileBuildOnly, Evidence: "TinyGo object compiled; device execution unverified"}
}

func conciseEvidence(output string) string {
	const limit = 240
	evidence := strings.Join(strings.Fields(output), " ")
	if evidence == "" {
		return "tool rejected the probe without diagnostics"
	}
	if len(evidence) > limit {
		return evidence[:limit-3] + "..."
	}
	return evidence
}

func profileProbes() []profileProbe {
	return []profileProbe{
		{name: "baseline", source: "package main\nfunc main() {}\n"},
		{name: "goroutine", source: "package main\nfunc main() { go func() {}() }\n"},
		{name: "channel", source: "package main\nfunc main() { c := make(chan int, 1); c <- 1; <-c }\n"},
		{name: "select", source: "package main\nfunc main() { c := make(chan int, 1); select { case c <- 1: default: } }\n"},
		{name: "defer", source: "package main\nfunc main() { defer func() {}() }\n"},
		{name: "panic", source: "package main\nfunc main() { panic(\"profile probe\") }\n"},
		{name: "recover", source: "package main\nfunc main() { defer func() { _ = recover() }() }\n"},
		{name: "reflection", source: "package main\nimport \"reflect\"\nfunc main() { _ = reflect.ValueOf(1).Interface() }\n"},
		{name: "finalizer", source: "package main\nimport \"runtime\"\ntype value struct{}\nfunc main() { runtime.SetFinalizer(&value{}, func(*value) {}) }\n"},
		{name: "cgo", source: "package main\n/* int answer(void) { return 42; } */\nimport \"C\"\nfunc main() { _ = C.answer() }\n"},
		{name: "runtime-gc", source: "package main\nimport \"runtime\"\nfunc main() { runtime.GC() }\n"},
		{name: "stdlib-time", source: "package main\nimport \"time\"\nfunc main() { _ = time.Second.String() }\n"},
		{name: "stdlib-fmt", source: "package main\nimport \"fmt\"\nfunc main() { _ = fmt.Sprint(42) }\n"},
		{name: "stdlib-json", source: "package main\nimport \"encoding/json\"\nfunc main() { _, _ = json.Marshal(struct{ Value int }{42}) }\n"},
	}
}
