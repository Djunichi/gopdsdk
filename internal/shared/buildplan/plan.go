// Package buildplan models inspectable cross-platform Playdate build commands.
package buildplan

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// Target identifies a Playdate artifact target.
type Target string

const (
	// Simulator targets the host Playdate Simulator.
	Simulator Target = "simulator"
	// Device targets physical Playdate hardware.
	Device Target = "device"

	deviceLinkerFlags = "-Wl,--gc-sections,--emit-relocs," +
		"--defsym,__exidx_start=0,--defsym,__exidx_end=0," +
		"--defsym,_sbss=__bss_start__,--defsym,_ebss=__bss_end__," +
		"--defsym,_sdata=__data_start__,--defsym,_edata=__data_end__,--defsym,_sidata=__etext," +
		"--defsym,_heap_start=__bss_end__,--defsym,_heap_end=__bss_end__," +
		"--defsym,_globals_start=__data_start__,--defsym,_globals_end=__bss_end__,--defsym,_stack_top=__bss_end__"
)

// Command is one process invocation with structured arguments.
type Command struct {
	Purpose     string
	Executable  string
	Args        []string
	Directory   string
	Environment []string
}

// Retention defines whether an artifact survives plan execution.
type Retention string

const (
	// Cleanup removes an owned temporary artifact when execution ends.
	Cleanup Retention = "cleanup"
	// Preserve keeps a published artifact after successful execution.
	Preserve Retention = "preserve"
)

// Artifact records a path and its ownership policy.
type Artifact struct {
	Name      string
	Path      string
	Retention Retention
}

// Plan describes a build without executing it.
type Plan struct {
	Target      Target
	Application string
	SDKPath     string
	Output      string
	Commands    []Command
	Artifacts   []Artifact
}

// New returns a deterministic semantic plan using portable workspace tokens.
func New(target Target, application, sdkPath, output string) (Plan, error) {
	if application == "" {
		application = "."
	}
	if sdkPath == "" {
		sdkPath = "${PLAYDATE_SDK_PATH}"
	}
	if output == "" {
		output = "build/${PACKAGE}.pdx"
	}
	plan := Plan{Target: target, Application: application, SDKPath: sdkPath, Output: output}
	plan.Artifacts = []Artifact{
		{Name: "workspace", Path: "${WORK}", Retention: Cleanup},
		{Name: "application package", Path: output, Retention: Preserve},
	}
	switch target {
	case Simulator:
		plan.Commands = simulatorCommands(sdkPath, output)
	case Device:
		plan.Commands = deviceCommands(sdkPath, output)
	default:
		return Plan{}, fmt.Errorf("unknown build target %q", target)
	}
	return plan, nil
}

func simulatorCommands(sdkPath, output string) []Command {
	return []Command{
		{Purpose: "compile Simulator shared library", Executable: "go", Args: []string{"build", "-buildmode=c-shared", "-o", "${WORK}/Source/pdex.${HOST_LIBRARY}", "."}, Directory: "${WORK}", Environment: []string{"CGO_ENABLED=1", "CC=${CC}", "CGO_CFLAGS=-DTARGET_EXTENSION=1 -DTARGET_SIMULATOR=1"}},
		{Purpose: "package Simulator application", Executable: "${PDC}", Args: []string{"-sdkpath", sdkPath, "${WORK}/Source", "${PACKAGE_OUTPUT}"}},
	}
}

// Resolve substitutes runtime tokens without assembling shell commands.
func Resolve(plan Plan, tokens map[string]string) Plan {
	resolved := plan
	resolved.Commands = make([]Command, len(plan.Commands))
	resolved.Artifacts = make([]Artifact, len(plan.Artifacts))
	for index, artifact := range plan.Artifacts {
		resolved.Artifacts[index] = artifact
		resolved.Artifacts[index].Path = resolveValue(artifact.Path, tokens)
	}
	for index, command := range plan.Commands {
		resolved.Commands[index] = command
		resolved.Commands[index].Executable = resolveValue(command.Executable, tokens)
		resolved.Commands[index].Directory = resolveValue(command.Directory, tokens)
		resolved.Commands[index].Args = resolveValues(command.Args, tokens)
		resolved.Commands[index].Environment = resolveValues(command.Environment, tokens)
	}
	return resolved
}

// CleanupPaths returns validated absolute paths owned by the resolved plan.
func CleanupPaths(plan Plan) ([]string, error) {
	var paths []string
	for _, artifact := range plan.Artifacts {
		if artifact.Retention != Cleanup {
			continue
		}
		path := filepath.Clean(artifact.Path)
		if strings.Contains(path, "${") || !filepath.IsAbs(path) || filepath.Dir(path) == path {
			return nil, fmt.Errorf("unsafe cleanup path for artifact %q: %q", artifact.Name, artifact.Path)
		}
		paths = append(paths, path)
	}
	return paths, nil
}

// BindExecutables replaces discovered tool names while preserving command arguments.
func BindExecutables(plan Plan, tools map[string]string) Plan {
	bound := plan
	bound.Commands = append([]Command(nil), plan.Commands...)
	for index := range bound.Commands {
		if executable, ok := tools[bound.Commands[index].Executable]; ok {
			bound.Commands[index].Executable = executable
		}
	}
	return bound
}

func resolveValues(values []string, tokens map[string]string) []string {
	resolved := make([]string, len(values))
	for index, value := range values {
		resolved[index] = resolveValue(value, tokens)
	}
	return resolved
}

func resolveValue(value string, tokens map[string]string) string {
	for token, replacement := range tokens {
		value = strings.ReplaceAll(value, token, replacement)
	}
	return value
}

func deviceCommands(sdkPath, output string) []Command {
	compileFlags := []string{"-c", "-mthumb", "-mcpu=cortex-m7", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16", "-O2", "-ffunction-sections", "-fdata-sections"}
	apiFlags := []string{"-DTARGET_PLAYDATE=1", "-DTARGET_EXTENSION=1", "-I", sdkFile(sdkPath, "C_API")}
	return []Command{
		{Purpose: "resolve Go module graph", Executable: "go", Args: []string{"mod", "tidy"}, Directory: "${WORK}"},
		{Purpose: "compile TinyGo PIC object", Executable: "tinygo", Args: []string{"build", "-target", "${WORK}/playdate.json", "-scheduler", "none", "-gc", "none", "-panic", "trap", "-opt", "0", "-o", "${WORK}/probe.o", "."}, Directory: "${WORK}"},
		{Purpose: "expose TinyGo runtime bootstrap symbol", Executable: "arm-none-eabi-objcopy", Args: []string{"--globalize-symbol=runtime.run", "${WORK}/probe.o"}, Directory: "${WORK}"},
		{Purpose: "compile official Playdate setup", Executable: "arm-none-eabi-gcc", Args: append(append(append([]string{}, compileFlags...), apiFlags...), sdkFile(sdkPath, "C_API", "buildsupport", "setup.c"), "-o", "${WORK}/setup.o"), Directory: "${WORK}"},
		{Purpose: "compile TinyGo runtime adapter", Executable: "arm-none-eabi-gcc", Args: append(append([]string{}, compileFlags...), "${WORK}/adapter.S", "-o", "${WORK}/adapter.o"), Directory: "${WORK}"},
		{Purpose: "compile TinyGo runtime bootstrap", Executable: "arm-none-eabi-gcc", Args: append(append(append([]string{}, compileFlags...), apiFlags...), "${WORK}/bootstrap.c", "-o", "${WORK}/bootstrap.o"), Directory: "${WORK}"},
		{Purpose: "link Playdate device ELF", Executable: "arm-none-eabi-gcc", Args: []string{"${WORK}/probe.o", "${WORK}/setup.o", "${WORK}/adapter.o", "${WORK}/bootstrap.o", "-nostartfiles", "-mthumb", "-mcpu=cortex-m7", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16", "-T", sdkFile(sdkPath, "C_API", "buildsupport", "link_map.ld"), deviceLinkerFlags, "-o", "${WORK}/pdex.elf"}, Directory: "${WORK}"},
		{Purpose: "inspect ARM object", Executable: "arm-none-eabi-readelf", Args: []string{"-h", "-A", "-s", "-r", "${WORK}/pdex.elf"}, Directory: "${WORK}"},
		{Purpose: "inspect unresolved ELF symbols", Executable: "arm-none-eabi-nm", Args: []string{"-u", "${WORK}/pdex.elf"}, Directory: "${WORK}"},
		{Purpose: "inspect linked ELF symbols", Executable: "arm-none-eabi-nm", Args: []string{"${WORK}/pdex.elf"}, Directory: "${WORK}"},
		{Purpose: "package device application", Executable: sdkTool(sdkPath, "pdc"), Args: []string{"-sdkpath", sdkPath, "-q", "${WORK}/Source", "${PACKAGE_OUTPUT}"}, Directory: "${WORK}"},
	}
}

func sdkTool(sdkPath, name string) string {
	return sdkFile(sdkPath, "bin", name)
}

func sdkFile(sdkPath string, parts ...string) string {
	return filepath.ToSlash(filepath.Join(append([]string{sdkPath}, parts...)...))
}

// Write renders a stable human-readable plan.
func Write(writer io.Writer, plan Plan) error {
	if _, err := fmt.Fprintf(writer, "Build plan\nTarget:      %s\nApplication: %s\nSDK:         %s\nOutput:      %s\nArtifacts:\n", plan.Target, plan.Application, plan.SDKPath, plan.Output); err != nil {
		return err
	}
	for _, artifact := range plan.Artifacts {
		if _, err := fmt.Fprintf(writer, "  - %s: %s (%s)\n", artifact.Name, artifact.Path, artifact.Retention); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(writer, "Commands:"); err != nil {
		return err
	}
	for index, command := range plan.Commands {
		parts := append([]string{command.Executable}, command.Args...)
		for item := range parts {
			parts[item] = strconv.Quote(parts[item])
		}
		if _, err := fmt.Fprintf(writer, "  %d. %s\n     %s\n", index+1, command.Purpose, strings.Join(parts, " ")); err != nil {
			return err
		}
		if command.Directory != "" {
			if _, err := fmt.Fprintf(writer, "     directory: %s\n", command.Directory); err != nil {
				return err
			}
		}
		for _, variable := range command.Environment {
			if _, err := fmt.Fprintf(writer, "     env: %s\n", variable); err != nil {
				return err
			}
		}
	}
	return nil
}
