// Package deviceprobe verifies the TinyGo ARM object compilation stage.
package deviceprobe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Result records the verified device toolchain stage.
type Result struct {
	TinyGo  string
	GCC     string
	Format  string
	Export  string
	Pending string
}

// Probe compiles and inspects a minimal Cortex-M7 Go object.
func Probe(ctx context.Context) (Result, error) {
	tinygo, err := exec.LookPath("tinygo")
	if err != nil {
		return Result{}, fmt.Errorf("find tinygo: %w", err)
	}
	gcc, err := exec.LookPath("arm-none-eabi-gcc")
	if err != nil {
		return Result{}, fmt.Errorf("find arm-none-eabi-gcc: %w", err)
	}
	readelf, err := exec.LookPath("arm-none-eabi-readelf")
	if err != nil {
		return Result{}, fmt.Errorf("find arm-none-eabi-readelf: %w", err)
	}
	workDir, err := os.MkdirTemp("", "gopdsdk-device-probe-")
	if err != nil {
		return Result{}, fmt.Errorf("create device probe directory: %w", err)
	}
	defer os.RemoveAll(workDir)
	for _, file := range []struct {
		name     string
		contents string
	}{
		{name: "go.mod", contents: fmt.Sprintf("module sdk.gopdsdk/deviceprobe\n\ngo %s\n", strings.TrimPrefix(runtime.Version(), "go"))},
		{name: "main.go", contents: probeSource},
	} {
		if err := os.WriteFile(filepath.Join(workDir, file.name), []byte(file.contents), 0o644); err != nil {
			return Result{}, fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	objectPath := filepath.Join(workDir, "probe.o")
	command := exec.CommandContext(ctx, tinygo, "build", "-target", "nucleo-f722ze", "-scheduler", "none", "-gc", "leaking", "-panic", "trap", "-o", objectPath, ".")
	command.Dir = workDir
	if output, runErr := command.CombinedOutput(); runErr != nil {
		return Result{}, commandError("compile TinyGo Cortex-M7 object", runErr, output)
	}
	command = exec.CommandContext(ctx, readelf, "-h", "-s", objectPath)
	output, runErr := command.CombinedOutput()
	if runErr != nil {
		return Result{}, commandError("inspect ARM object", runErr, output)
	}
	inspection := string(output)
	for _, check := range []struct {
		description string
		required    string
	}{
		{description: "ELF class", required: "Class:                             ELF32"},
		{description: "ELF type", required: "Type:                              REL (Relocatable file)"},
		{description: "machine", required: "Machine:                           ARM"},
		{description: "export", required: "eventHandler"},
	} {
		if !strings.Contains(inspection, check.required) {
			return Result{}, fmt.Errorf("inspect ARM object: %s was not verified", check.description)
		}
	}
	return Result{
		TinyGo:  firstLine(runVersion(ctx, tinygo, "version")),
		GCC:     firstLine(runVersion(ctx, gcc, "--version")),
		Format:  "ELF32 ARM relocatable object",
		Export:  "eventHandler",
		Pending: "Playdate runtime adaptation and pdex.elf link",
	}, nil
}

func runVersion(ctx context.Context, executable, argument string) string {
	output, err := exec.CommandContext(ctx, executable, argument).CombinedOutput()
	if err != nil {
		return executable
	}
	return string(output)
}

func firstLine(value string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(value), "\n")
	return strings.TrimSpace(line)
}

func commandError(action string, err error, output []byte) error {
	detail := strings.TrimSpace(string(output))
	if detail == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, detail)
}

const probeSource = `package main

//export eventHandler
func eventHandler(uintptr, int32, uint32) int32 { return 0 }

func main() {}
`
