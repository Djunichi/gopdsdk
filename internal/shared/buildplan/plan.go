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
)

// Command is one process invocation with structured arguments.
type Command struct {
	Purpose    string
	Executable string
	Args       []string
	Directory  string
}

// Plan describes a build without executing it.
type Plan struct {
	Target      Target
	Application string
	SDKPath     string
	Output      string
	Commands    []Command
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
		{Purpose: "compile Simulator shared library", Executable: "go", Args: []string{"build", "-buildmode=c-shared", "-o", "${WORK}/Source/pdex.${HOST_LIBRARY}", "."}, Directory: "${WORK}"},
		{Purpose: "package Simulator application", Executable: sdkTool(sdkPath, "pdc"), Args: []string{"-sdkpath", sdkPath, "${WORK}/Source", output}},
	}
}

func deviceCommands(sdkPath, output string) []Command {
	compileFlags := []string{"-c", "-mthumb", "-mcpu=cortex-m7", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16"}
	return []Command{
		{Purpose: "compile TinyGo PIC object", Executable: "tinygo", Args: []string{"build", "-target", "${WORK}/playdate.json", "-scheduler", "none", "-gc", "none", "-panic", "trap", "-opt", "0", "-o", "${WORK}/probe.o", "."}, Directory: "${WORK}"},
		{Purpose: "expose TinyGo runtime bootstrap symbol", Executable: "arm-none-eabi-objcopy", Args: []string{"--globalize-symbol=runtime.run", "${WORK}/probe.o"}},
		{Purpose: "compile official Playdate setup", Executable: "arm-none-eabi-gcc", Args: append(append([]string{}, compileFlags...), sdkFile(sdkPath, "C_API", "buildsupport", "setup.c"), "-o", "${WORK}/setup.o")},
		{Purpose: "compile runtime adapter", Executable: "arm-none-eabi-gcc", Args: append(append([]string{}, compileFlags...), "${WORK}/adapter.S", "-o", "${WORK}/adapter.o")},
		{Purpose: "compile runtime bootstrap", Executable: "arm-none-eabi-gcc", Args: append(append([]string{}, compileFlags...), "${WORK}/bootstrap.c", "-o", "${WORK}/bootstrap.o")},
		{Purpose: "link Playdate device ELF", Executable: "arm-none-eabi-gcc", Args: []string{"${WORK}/probe.o", "${WORK}/setup.o", "${WORK}/adapter.o", "${WORK}/bootstrap.o", "-nostartfiles", "-mthumb", "-mcpu=cortex-m7", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16", "-T", sdkFile(sdkPath, "C_API", "buildsupport", "link_map.ld"), "-o", "${WORK}/pdex.elf"}},
		{Purpose: "inspect ABI, symbols, and relocations", Executable: "arm-none-eabi-readelf", Args: []string{"-h", "-A", "-s", "-r", "${WORK}/pdex.elf"}},
		{Purpose: "inspect unresolved symbols", Executable: "arm-none-eabi-nm", Args: []string{"-u", "${WORK}/pdex.elf"}},
		{Purpose: "package device application", Executable: sdkTool(sdkPath, "pdc"), Args: []string{"-sdkpath", sdkPath, "${WORK}/Source", output}},
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
	if _, err := fmt.Fprintf(writer, "Build plan\nTarget:      %s\nApplication: %s\nSDK:         %s\nOutput:      %s\nCommands:\n", plan.Target, plan.Application, plan.SDKPath, plan.Output); err != nil {
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
	}
	return nil
}
