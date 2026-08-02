// Package deviceprobe verifies TinyGo ARM linking and Playdate packaging.
package deviceprobe

import (
	"context"
	"debug/elf"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Djunichi/gopdsdk/internal/shared/buildplan"
	"github.com/Djunichi/gopdsdk/internal/shared/gomodule"
	"github.com/Djunichi/gopdsdk/internal/shared/pdxsource"
)

// Result records the verified device toolchain stage.
type Result struct {
	TinyGo  string
	GCC     string
	Format  string
	Export  string
	Package string
	Output  string
	Deploy  string
	Run     string
	Pending string
	Metrics Metrics
}

// Metrics records reproducible linked and packaged device sizes in bytes.
type Metrics struct {
	StaticRAM uint64
	ELF       int64
	PDX       int64
}

// Config identifies the official Playdate SDK used for the link stage.
type Config struct {
	SDKPath      string
	Application  string
	Output       string
	Replace      bool
	Persist      bool
	Install      bool
	Run          bool
	ArtifactsDir string
	Memory       buildplan.DeviceMemoryStrategy
}

// Probe compiles and links a structural Playdate device ELF.
func Probe(ctx context.Context, config Config) (Result, error) {
	if config.SDKPath == "" {
		return Result{}, fmt.Errorf("Playdate SDK path is required")
	}
	module, err := gomodule.Locate(ctx)
	if err != nil {
		return Result{}, err
	}
	if config.Application == "" {
		config.Application = module.Path + "/examples/hello"
	}
	app, err := inspectApplication(ctx, config.Application)
	if err != nil {
		return Result{}, err
	}
	if app.Name == "main" || app.Module == nil {
		return Result{}, fmt.Errorf("inspect application package: %s must be an importable package in a Go module", app.ImportPath)
	}
	pdxInfo, err := os.ReadFile(filepath.Join(app.Dir, "pdxinfo"))
	if err != nil {
		return Result{}, fmt.Errorf("read application pdxinfo: %w", err)
	}
	if config.Persist && config.Output == "" {
		config.Output = filepath.Join("build", app.Name+".pdx")
	}
	sdkPath, err := filepath.Abs(filepath.Clean(config.SDKPath))
	if err != nil {
		return Result{}, fmt.Errorf("resolve Playdate SDK path: %w", err)
	}
	setupSource := filepath.Join(sdkPath, "C_API", "buildsupport", "setup.c")
	linkerScript := filepath.Join(sdkPath, "C_API", "buildsupport", "link_map.ld")
	pdcName := "pdc"
	if runtime.GOOS == "windows" {
		pdcName += ".exe"
	}
	pdc := filepath.Join(sdkPath, "bin", pdcName)
	for _, path := range []string{filepath.Join(sdkPath, "C_API", "pd_api.h"), setupSource, linkerScript, pdc} {
		if info, statErr := os.Stat(path); statErr != nil || info.IsDir() {
			return Result{}, fmt.Errorf("required file %s is unavailable", path)
		}
	}
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
	nm, err := exec.LookPath("arm-none-eabi-nm")
	if err != nil {
		return Result{}, fmt.Errorf("find arm-none-eabi-nm: %w", err)
	}
	objcopy, err := exec.LookPath("arm-none-eabi-objcopy")
	if err != nil {
		return Result{}, fmt.Errorf("find arm-none-eabi-objcopy: %w", err)
	}
	workDir, err := os.MkdirTemp("", "gopdsdk-device-probe-")
	if err != nil {
		return Result{}, fmt.Errorf("create device probe directory: %w", err)
	}
	elfPath := filepath.Join(workDir, "pdex.elf")
	pdxName := app.Name + ".pdx"
	pdxPath := filepath.Join(workDir, pdxName)
	if config.Memory == "" {
		config.Memory = buildplan.DeviceMemoryConservative
	}
	plan, err := buildplan.NewDevice(config.Application, sdkPath, config.Output, config.Memory)
	if err != nil {
		_ = os.RemoveAll(workDir)
		return Result{}, err
	}
	plan = buildplan.Resolve(plan, map[string]string{"${WORK}": workDir, "${PACKAGE_OUTPUT}": pdxPath})
	plan = buildplan.BindExecutables(plan, map[string]string{
		"tinygo": tinygo, "arm-none-eabi-objcopy": objcopy, "arm-none-eabi-gcc": gcc,
		"arm-none-eabi-readelf": readelf, "arm-none-eabi-nm": nm, plan.Commands[10].Executable: pdc,
	})
	cleanupPaths, err := buildplan.CleanupPaths(plan)
	if err != nil {
		_ = os.RemoveAll(workDir)
		return Result{}, err
	}
	defer cleanupArtifacts(cleanupPaths)
	for _, file := range []struct {
		name     string
		contents string
	}{
		{name: "go.mod", contents: renderDeviceGoMod(module, app)},
		{name: "main.go", contents: renderProbeSource(module.Path, app.ImportPath)},
		{name: "playdate.json", contents: targetSource},
		{name: "adapter.S", contents: adapterSource},
	} {
		if err := os.WriteFile(filepath.Join(workDir, file.name), []byte(file.contents), 0o644); err != nil {
			return Result{}, fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	for index, planned := range plan.Commands[:7] {
		if _, err := runPlannedCommand(ctx, planned); err != nil {
			return Result{}, err
		}
		if index == 2 {
			bootstrap := bootstrapSource
			if config.Memory == buildplan.DeviceMemoryConservative {
				bootstrap = conservativeBootstrapSource
			}
			if err := os.WriteFile(filepath.Join(workDir, "bootstrap.c"), []byte(bootstrap), 0o644); err != nil {
				return Result{}, fmt.Errorf("write bootstrap.c: %w", err)
			}
		}
	}
	inspectionOutput, err := runPlannedCommand(ctx, plan.Commands[7])
	if err != nil {
		return Result{}, err
	}
	inspection := inspectionOutput
	for _, check := range []struct {
		description string
		required    string
	}{
		{description: "ELF class", required: "Class:                             ELF32"},
		{description: "ELF type", required: "Type:                              EXEC (Executable file)"},
		{description: "machine", required: "Machine:                           ARM"},
		{description: "export", required: "eventHandler"},
		{description: "entry point", required: "eventHandlerShim"},
		{description: "Go event handler", required: "goEventHandler"},
		{description: "Go update handler", required: "goUpdate"},
		{description: "runtime run", required: "runtime.run"},
		{description: "hard-float ABI", required: "Tag_ABI_VFP_args: VFP registers"},
	} {
		if !strings.Contains(inspection, check.required) {
			return Result{}, fmt.Errorf("inspect ARM object: %s was not verified", check.description)
		}
	}
	if config.Memory == buildplan.DeviceMemoryConservative {
		for _, check := range []struct{ description, required string }{
			{description: "bounded TinyGo heap", required: "playdateRuntimeHeap"},
			{description: "heap start", required: "_heap_start"},
			{description: "heap end", required: "_heap_end"},
			{description: "runtime stack boundary", required: "runtime.stackTop"},
			{description: "shadowed Cortex-M system control pointer", required: "playdateRuntimeSCB"},
		} {
			if !strings.Contains(inspection, check.required) {
				return Result{}, fmt.Errorf("inspect ARM object: %s was not verified", check.description)
			}
		}
	}
	for _, forbidden := range []string{"R_ARM_THM_MOVW_AB", "R_ARM_THM_MOVT_AB"} {
		if strings.Contains(inspection, forbidden) {
			return Result{}, fmt.Errorf("inspect ARM object: unsupported Playdate relocation %s remains", forbidden)
		}
	}
	undefinedOutput, err := runPlannedCommand(ctx, plan.Commands[8])
	if err != nil {
		return Result{}, err
	}
	if unresolved := strongUndefinedSymbols(undefinedOutput); len(unresolved) != 0 {
		return Result{}, fmt.Errorf("inspect unresolved ELF symbols: %s", strings.Join(unresolved, ", "))
	}
	symbolOutput, err := runPlannedCommand(ctx, plan.Commands[9])
	if err != nil {
		return Result{}, err
	}
	lowerSymbols := strings.ToLower(symbolOutput)
	for _, forbidden := range []string{"stm32", "initclk", "machine.tim"} {
		if strings.Contains(lowerSymbols, forbidden) {
			return Result{}, fmt.Errorf("inspect linked ELF symbols: board-specific symbol %q remains", forbidden)
		}
	}
	if unsupported := unsupportedRuntimeSymbols(symbolOutput, config.Memory); len(unsupported) != 0 {
		return Result{}, fmt.Errorf("inspect linked ELF symbols: unsupported TinyGo runtime symbols remain: %s", strings.Join(unsupported, ", "))
	}
	if config.Memory == buildplan.DeviceMemoryConservative {
		if err := validateConservativeHeapSymbols(symbolOutput); err != nil {
			return Result{}, fmt.Errorf("inspect conservative GC heap: %w", err)
		}
	}
	sourceDir := filepath.Join(workDir, "Source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create device package source: %w", err)
	}
	if err := pdxsource.Stage(app.Dir, sourceDir); err != nil {
		return Result{}, err
	}
	packagedELFSource := filepath.Join(sourceDir, "pdex.elf")
	if err := copyFile(elfPath, packagedELFSource); err != nil {
		return Result{}, fmt.Errorf("stage device ELF: %w", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "pdxinfo"), pdxInfo, 0o644); err != nil {
		return Result{}, fmt.Errorf("write device pdxinfo: %w", err)
	}
	if _, err := runPlannedCommand(ctx, plan.Commands[10]); err != nil {
		return Result{}, err
	}
	packagedBinary := filepath.Join(pdxPath, "pdex.bin")
	if err := requireNonEmptyFile(packagedBinary); err != nil {
		return Result{}, fmt.Errorf("verify packaged device binary: %w", err)
	}
	if info, statErr := os.Stat(filepath.Join(pdxPath, "pdxinfo")); statErr != nil || info.IsDir() {
		return Result{}, fmt.Errorf("verify packaged pdxinfo: pdc did not create a file")
	}
	metrics, err := measureDeviceArtifact(elfPath, pdxPath)
	if err != nil {
		return Result{}, err
	}
	if config.ArtifactsDir != "" {
		artifactsDir, err := filepath.Abs(filepath.Clean(config.ArtifactsDir))
		if err != nil {
			return Result{}, fmt.Errorf("resolve artifacts directory: %w", err)
		}
		if err := os.MkdirAll(artifactsDir, 0o755); err != nil {
			return Result{}, fmt.Errorf("create artifacts directory: %w", err)
		}
		if err := copyFile(elfPath, filepath.Join(artifactsDir, "pdex.elf")); err != nil {
			return Result{}, fmt.Errorf("save diagnostic ELF: %w", err)
		}
	}
	artifactOutput := "temporary package"
	if config.Output != "" {
		outputPath, err := filepath.Abs(filepath.Clean(config.Output))
		if err != nil {
			return Result{}, fmt.Errorf("resolve output path: %w", err)
		}
		if info, statErr := os.Stat(outputPath); statErr == nil {
			if !config.Replace {
				return Result{}, fmt.Errorf("output already exists: %s", outputPath)
			}
			if !info.IsDir() {
				return Result{}, fmt.Errorf("output path is not a directory: %s", outputPath)
			}
			if err := os.RemoveAll(outputPath); err != nil {
				return Result{}, fmt.Errorf("replace output: %w", err)
			}
		} else if !os.IsNotExist(statErr) {
			return Result{}, fmt.Errorf("inspect output path: %w", statErr)
		}
		if err := copyDirectory(pdxPath, outputPath); err != nil {
			return Result{}, fmt.Errorf("write output: %w", err)
		}
		artifactOutput = outputPath
	}
	deployment := "not requested"
	execution := "not requested"
	pending := "device deployment and hardware execution; allocator is non-collecting"
	if config.Memory == buildplan.DeviceMemoryConservative {
		pending = "device deployment, hardware execution, and conservative-GC soak"
	}
	if config.Install || config.Run {
		pdutilName := "pdutil"
		if runtime.GOOS == "windows" {
			pdutilName += ".exe"
		}
		pdutil := filepath.Join(sdkPath, "bin", pdutilName)
		if info, statErr := os.Stat(pdutil); statErr != nil || info.IsDir() {
			return Result{}, fmt.Errorf("required file %s is unavailable", pdutil)
		}
		installOutput, err := runCommandOutput(ctx, workDir, "install device probe", pdutil, "install", pdxPath)
		if err != nil {
			return Result{}, err
		}
		deployment = summarizeOutput(installOutput)
		pending = "hardware execution; allocator is non-collecting"
		if config.Memory == buildplan.DeviceMemoryConservative {
			pending = "hardware execution and conservative-GC soak"
		}
		if config.Run {
			runOutput, err := runDeviceProbe(ctx, workDir, pdutil, "/Games/"+pdxName)
			if err != nil {
				return Result{}, err
			}
			execution = summarizeRunOutput(runOutput)
			pending = "allocator is non-collecting"
			if config.Memory == buildplan.DeviceMemoryConservative {
				pending = "conservative-GC soak and memory-growth measurement"
			}
		}
	}
	return Result{
		TinyGo:  firstLine(runVersion(ctx, tinygo, "version")),
		GCC:     firstLine(runVersion(ctx, gcc, "--version")),
		Format:  "ELF32 ARM executable, hard-float",
		Export:  "eventHandler",
		Package: app.ImportPath,
		Output:  artifactOutput,
		Deploy:  deployment,
		Run:     execution,
		Pending: pending,
		Metrics: metrics,
	}, nil
}

func measureDeviceArtifact(elfPath, pdxPath string) (Metrics, error) {
	elfInfo, err := os.Stat(elfPath)
	if err != nil {
		return Metrics{}, fmt.Errorf("measure device ELF: %w", err)
	}
	image, err := elf.Open(elfPath)
	if err != nil {
		return Metrics{}, fmt.Errorf("measure static RAM: %w", err)
	}
	defer image.Close()
	var staticRAM uint64
	for _, section := range image.Sections {
		if section.Flags&elf.SHF_ALLOC != 0 && section.Flags&elf.SHF_WRITE != 0 {
			staticRAM += section.Size
		}
	}
	pdxSize, err := directoryFileSize(pdxPath)
	if err != nil {
		return Metrics{}, fmt.Errorf("measure device PDX: %w", err)
	}
	return Metrics{StaticRAM: staticRAM, ELF: elfInfo.Size(), PDX: pdxSize}, nil
}

func directoryFileSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func cleanupArtifacts(paths []string) {
	for _, path := range paths {
		_ = os.RemoveAll(path)
	}
}

func runDeviceProbe(ctx context.Context, directory, pdutil, devicePath string) (string, error) {
	const attempts = 10
	for attempt := 1; attempt <= attempts; attempt++ {
		output, err := runCommandOutput(ctx, directory, "run device probe", pdutil, "run", devicePath)
		if err == nil {
			return output, nil
		}
		if !strings.Contains(err.Error(), "No Playdate device detected") || attempt == attempts {
			return "", err
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("run device probe: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}
	return "", fmt.Errorf("run device probe: retry limit reached")
}

func copyFile(source, destination string) error {
	contents, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(destination, contents, 0o644)
}

func copyDirectory(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func requireNonEmptyFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", path)
	}
	if info.Size() == 0 {
		return fmt.Errorf("%s is empty", path)
	}
	return nil
}

func runPlannedCommand(ctx context.Context, planned buildplan.Command) (string, error) {
	return runCommandOutput(ctx, planned.Directory, planned.Purpose, planned.Executable, planned.Args...)
}

func runCommandOutput(ctx context.Context, directory, action, executable string, args ...string) (string, error) {
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = directory
	output, err := command.CombinedOutput()
	if err != nil {
		return "", commandError(action, err, output)
	}
	return strings.TrimSpace(string(output)), nil
}

func summarizeOutput(output string) string {
	if summary := strings.Join(strings.Fields(output), " "); summary != "" {
		return summary
	}
	return "installed by pdutil"
}

func summarizeRunOutput(output string) string {
	if summary := strings.Join(strings.Fields(output), " "); summary != "" {
		return summary
	}
	return "launch command sent by pdutil"
}

func strongUndefinedSymbols(output string) []string {
	var symbols []string
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[len(fields)-2] == "U" {
			symbols = append(symbols, fields[len(fields)-1])
		}
	}
	return symbols
}

func unsupportedRuntimeSymbols(output string, memory buildplan.DeviceMemoryStrategy) []string {
	var symbols []string
	for _, forbidden := range []string{
		"runtime.setupDeferFrame",
		"runtime._recover",
		"runtime.chan",
		"runtime.SetFinalizer",
		"runtime.setFinalizer",
		" reflect.",
	} {
		if strings.Contains(output, forbidden) {
			symbols = append(symbols, strings.TrimSpace(forbidden))
		}
	}
	if memory != buildplan.DeviceMemoryConservative && strings.Contains(output, "runtime/interrupt.In") {
		symbols = append(symbols, "runtime/interrupt.In")
	}
	return symbols
}

const conservativeHeapSize = uint64(256 * 1024)

func validateConservativeHeapSymbols(output string) error {
	addresses := make(map[string]uint64)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		address, err := strconv.ParseUint(fields[0], 16, 64)
		if err == nil {
			addresses[fields[len(fields)-1]] = address
		}
	}
	required := []string{"_globals_end", "_heap_start", "_heap_end", "__bss_end__", "playdateRuntimeHeap", "playdateRuntimeSCB", "runtime.stackTop", "runtime.runtimePanicAt", "runtime/interrupt.In", "tinygo_scanCurrentStack"}
	for _, name := range required {
		if _, ok := addresses[name]; !ok {
			return fmt.Errorf("required symbol %s is missing", name)
		}
	}
	start := addresses["_heap_start"]
	end := addresses["_heap_end"]
	if start != addresses["_globals_end"] || start != addresses["playdateRuntimeHeap"] {
		return fmt.Errorf("heap start %#x does not match globals end %#x and storage %#x", start, addresses["_globals_end"], addresses["playdateRuntimeHeap"])
	}
	if start%16 != 0 {
		return fmt.Errorf("heap start %#x is not 16-byte aligned", start)
	}
	if end < start || end-start != conservativeHeapSize {
		return fmt.Errorf("heap range %#x..%#x is %d bytes, want %d", start, end, end-start, conservativeHeapSize)
	}
	if end > addresses["__bss_end__"] {
		return fmt.Errorf("heap end %#x exceeds BSS end %#x", end, addresses["__bss_end__"])
	}
	return nil
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

func renderProbeSource(modulePath, applicationImport string) string {
	return fmt.Sprintf(`package main

import (
	"unsafe"

	sdkRuntime %q
	sdkPlaydate %q
	app %q
)

//go:linkname bridgeClear bridgeClear
func bridgeClear()

//go:linkname bridgeDrawText bridgeDrawText
func bridgeDrawText(text *byte, length uintptr, x, y int32)
//go:linkname bridgeLoadFont bridgeLoadFont
func bridgeLoadFont(path *byte, message *uintptr) uintptr
//go:linkname bridgeSetFont bridgeSetFont
func bridgeSetFont(font uintptr)
//go:linkname bridgeTextWidth bridgeTextWidth
func bridgeTextWidth(font uintptr, text *byte, length uintptr) int32
//go:linkname bridgeFontHeight bridgeFontHeight
func bridgeFontHeight(font uintptr) int32
//go:linkname bridgeFreeFont bridgeFreeFont
func bridgeFreeFont(font uintptr)

//go:linkname bridgeCurrentTimeMilliseconds bridgeCurrentTimeMilliseconds
func bridgeCurrentTimeMilliseconds() uint32

//go:linkname bridgeButtons bridgeButtons
func bridgeButtons() uint32

//go:linkname bridgeCrankAngleBits bridgeCrankAngleBits
func bridgeCrankAngleBits() uint32

//go:linkname bridgeCrankDeltaBits bridgeCrankDeltaBits
func bridgeCrankDeltaBits() uint32

//go:linkname bridgeCrankDocked bridgeCrankDocked
func bridgeCrankDocked() int32

//go:linkname bridgeFrameDeltaBits bridgeFrameDeltaBits
func bridgeFrameDeltaBits() uint32

//go:linkname bridgeLoadBitmap bridgeLoadBitmap
func bridgeLoadBitmap(path *byte, message *uintptr) uintptr
//go:linkname bridgeLoadBitmapTable bridgeLoadBitmapTable
func bridgeLoadBitmapTable(path *byte, message *uintptr) uintptr
//go:linkname bridgeBitmapTableFrame bridgeBitmapTableFrame
func bridgeBitmapTableFrame(table uintptr, index int32) uintptr
//go:linkname bridgeFreeBitmapTable bridgeFreeBitmapTable
func bridgeFreeBitmapTable(table uintptr)
//go:linkname bridgeNewBitmap bridgeNewBitmap
func bridgeNewBitmap(width, height int32) uintptr
//go:linkname bridgeFreeBitmap bridgeFreeBitmap
func bridgeFreeBitmap(bitmap uintptr)
//go:linkname bridgeBitmapSize bridgeBitmapSize
func bridgeBitmapSize(bitmap uintptr, width, height *int32)
//go:linkname bridgeFillBitmap bridgeFillBitmap
func bridgeFillBitmap(bitmap uintptr, color int32)
//go:linkname bridgeDrawBitmap bridgeDrawBitmap
func bridgeDrawBitmap(bitmap uintptr, x, y int32)
//go:linkname bridgeDrawScaledBitmapBits bridgeDrawScaledBitmapBits
func bridgeDrawScaledBitmapBits(bitmap uintptr, x, y int32, scaleX, scaleY uint32)
//go:linkname bridgeDrawPrimitive bridgeDrawPrimitive
func bridgeDrawPrimitive(kind, x1, y1, x2, y2, x3, y3, lineWidth, solid int32, startAngle, endAngle uint32, pattern uintptr, patterned int32)
//go:linkname bridgeSetClipRect bridgeSetClipRect
func bridgeSetClipRect(x, y, width, height int32)
//go:linkname bridgeClearClipRect bridgeClearClipRect
func bridgeClearClipRect()
//go:linkname bridgeSetDrawOffset bridgeSetDrawOffset
func bridgeSetDrawOffset(dx, dy int32)
//go:linkname bridgeSetDrawMode bridgeSetDrawMode
func bridgeSetDrawMode(mode int32)
//go:linkname bridgeNewSprite bridgeNewSprite
func bridgeNewSprite() uintptr
//go:linkname bridgeFreeSprite bridgeFreeSprite
func bridgeFreeSprite(sprite uintptr)
//go:linkname bridgeSpriteSetBitmap bridgeSpriteSetBitmap
func bridgeSpriteSetBitmap(sprite, bitmap uintptr)
//go:linkname bridgeSpriteMoveToBits bridgeSpriteMoveToBits
func bridgeSpriteMoveToBits(sprite uintptr, x, y uint32)
//go:linkname bridgeSpriteMoveByBits bridgeSpriteMoveByBits
func bridgeSpriteMoveByBits(sprite uintptr, dx, dy uint32)
//go:linkname bridgeSpriteSetVisible bridgeSpriteSetVisible
func bridgeSpriteSetVisible(sprite uintptr, visible int32)
//go:linkname bridgeSpriteSetZIndex bridgeSpriteSetZIndex
func bridgeSpriteSetZIndex(sprite uintptr, z int32)
//go:linkname bridgeSpriteSetCollideRectBits bridgeSpriteSetCollideRectBits
func bridgeSpriteSetCollideRectBits(sprite uintptr, x, y, width, height uint32)
//go:linkname bridgeSpriteClearCollideRect bridgeSpriteClearCollideRect
func bridgeSpriteClearCollideRect(sprite uintptr)
//go:linkname bridgeSpriteSetTag bridgeSpriteSetTag
func bridgeSpriteSetTag(sprite uintptr, tag uint8)
//go:linkname bridgeSpriteMoveWithCollisionsBits bridgeSpriteMoveWithCollisionsBits
func bridgeSpriteMoveWithCollisionsBits(sprite uintptr, x, y uint32, actualX, actualY *uint32, count *int32) uintptr
//go:linkname bridgeCollisionValueBits bridgeCollisionValueBits
func bridgeCollisionValueBits(collisions uintptr, index, field int32) uint32
//go:linkname bridgeCollisionOther bridgeCollisionOther
func bridgeCollisionOther(collisions uintptr, index int32) uintptr
//go:linkname bridgeFreeList bridgeFreeList
func bridgeFreeList(list uintptr)
//go:linkname bridgeQuerySpritesAtPointBits bridgeQuerySpritesAtPointBits
func bridgeQuerySpritesAtPointBits(x, y uint32, count *int32) uintptr
//go:linkname bridgeQuerySpritesInRectBits bridgeQuerySpritesInRectBits
func bridgeQuerySpritesInRectBits(x, y, width, height uint32, count *int32) uintptr
//go:linkname bridgeSpriteListItem bridgeSpriteListItem
func bridgeSpriteListItem(list uintptr, index int32) uintptr
//go:linkname bridgeOverlappingSprites bridgeOverlappingSprites
func bridgeOverlappingSprites(sprite uintptr, count *int32) uintptr
//go:linkname bridgeSpriteAdd bridgeSpriteAdd
func bridgeSpriteAdd(sprite uintptr)
//go:linkname bridgeSpriteRemove bridgeSpriteRemove
func bridgeSpriteRemove(sprite uintptr)
//go:linkname bridgeUpdateAndDrawSprites bridgeUpdateAndDrawSprites
func bridgeUpdateAndDrawSprites()
//go:linkname bridgeLoadSoundEffect bridgeLoadSoundEffect
func bridgeLoadSoundEffect(path *byte) uintptr
//go:linkname bridgeSoundEffectPlay bridgeSoundEffectPlay
func bridgeSoundEffectPlay(effect uintptr) int32
//go:linkname bridgeSoundEffectStop bridgeSoundEffectStop
func bridgeSoundEffectStop(effect uintptr)
//go:linkname bridgeSoundEffectSetVolumeBits bridgeSoundEffectSetVolumeBits
func bridgeSoundEffectSetVolumeBits(effect uintptr, left, right uint32)
//go:linkname bridgeSoundEffectVolumeBits bridgeSoundEffectVolumeBits
func bridgeSoundEffectVolumeBits(effect uintptr, left, right *uint32)
//go:linkname bridgeSoundEffectIsPlaying bridgeSoundEffectIsPlaying
func bridgeSoundEffectIsPlaying(effect uintptr) int32
//go:linkname bridgeSoundEffectPause bridgeSoundEffectPause
func bridgeSoundEffectPause(effect uintptr, paused int32)
//go:linkname bridgeFreeSoundEffect bridgeFreeSoundEffect
func bridgeFreeSoundEffect(effect uintptr)
//go:linkname bridgeLoadFilePlayer bridgeLoadFilePlayer
func bridgeLoadFilePlayer(path *byte) uintptr
//go:linkname bridgeFilePlayerPlay bridgeFilePlayerPlay
func bridgeFilePlayerPlay(player uintptr) int32
//go:linkname bridgeFilePlayerStop bridgeFilePlayerStop
func bridgeFilePlayerStop(player uintptr)
//go:linkname bridgeFilePlayerSetVolumeBits bridgeFilePlayerSetVolumeBits
func bridgeFilePlayerSetVolumeBits(player uintptr, left, right uint32)
//go:linkname bridgeFilePlayerVolumeBits bridgeFilePlayerVolumeBits
func bridgeFilePlayerVolumeBits(player uintptr, left, right *uint32)
//go:linkname bridgeFilePlayerIsPlaying bridgeFilePlayerIsPlaying
func bridgeFilePlayerIsPlaying(player uintptr) int32
//go:linkname bridgeFilePlayerPause bridgeFilePlayerPause
func bridgeFilePlayerPause(player uintptr)
//go:linkname bridgeFreeFilePlayer bridgeFreeFilePlayer
func bridgeFreeFilePlayer(player uintptr)

func float32FromBits(bits uint32) float32 { return *(*float32)(unsafe.Pointer(&bits)) }

type playdateContext struct{}

var _ sdkPlaydate.PrimitiveGraphics = playdateContext{}
var _ sdkPlaydate.GraphicsState = playdateContext{}

func (playdateContext) Clear() { bridgeClear() }

func (playdateContext) DrawText(text string, x, y int) {
	bridgeDrawText(unsafe.StringData(text), uintptr(len(text)), int32(x), int32(y))
}

var fontDriver = sdkRuntime.FontDriver{
	TextWidth: func(handle uintptr, text string) int { return int(bridgeTextWidth(handle, unsafe.StringData(text), uintptr(len(text)))) },
	Height: func(handle uintptr) int { return int(bridgeFontHeight(handle)) },
	Free: bridgeFreeFont,
}
func (playdateContext) LoadFont(path string) (sdkPlaydate.Font, error) {
	terminated := path + "\x00"; var message uintptr
	handle := bridgeLoadFont(unsafe.StringData(terminated), &message)
	if handle == 0 { if message != 0 { return nil, sdkPlaydate.FontLoadError(cString(message)) }; return nil, sdkPlaydate.FontLoadError("unknown error") }
	return sdkRuntime.NewOwnedFont(handle, fontDriver), nil
}
func (playdateContext) DrawTextFont(font sdkPlaydate.Font, text string, x, y int) error {
	handle, err := sdkRuntime.FontHandle(font); if err != nil { return err }
	bridgeSetFont(handle); bridgeDrawText(unsafe.StringData(text), uintptr(len(text)), int32(x), int32(y)); return nil
}

func (playdateContext) CurrentTimeMilliseconds() uint32 {
	return bridgeCurrentTimeMilliseconds()
}

func (playdateContext) Input() sdkPlaydate.Input { return sdkPlaydate.Input{} }

var bitmapDriver = sdkRuntime.BitmapDriver{
	Dimensions: func(handle uintptr) (int, int) { var width, height int32; bridgeBitmapSize(handle, &width, &height); return int(width), int(height) },
	Fill: func(handle uintptr, color sdkPlaydate.Color) { bridgeFillBitmap(handle, int32(color)) },
	Free: bridgeFreeBitmap,
}
var bitmapTableDriver = sdkRuntime.BitmapTableDriver{
	Frame: func(table uintptr, index int) uintptr { return bridgeBitmapTableFrame(table, int32(index)) },
	Free: bridgeFreeBitmapTable,
}

func (playdateContext) LoadBitmap(path string) (sdkPlaydate.Bitmap, error) {
	terminated := path + "\x00"; var message uintptr
	handle := bridgeLoadBitmap(unsafe.StringData(terminated), &message)
	if handle == 0 { if message != 0 { return nil, sdkPlaydate.BitmapLoadError(cString(message)) }; return nil, sdkPlaydate.BitmapLoadError("unknown error") }
	return sdkRuntime.NewOwnedBitmap(handle, bitmapDriver), nil
}
func (playdateContext) LoadBitmapTable(path string) (sdkPlaydate.BitmapTable, error) {
	terminated := path + "\x00"; var message uintptr
	handle := bridgeLoadBitmapTable(unsafe.StringData(terminated), &message)
	if handle == 0 { if message != 0 { return nil, sdkPlaydate.BitmapLoadError(cString(message)) }; return nil, sdkPlaydate.BitmapLoadError("unknown error") }
	return sdkRuntime.NewOwnedBitmapTable(handle, bitmapTableDriver, bitmapDriver), nil
}
func cString(pointer uintptr) string { length := 0; for *(*byte)(unsafe.Pointer(pointer + uintptr(length))) != 0 { length++ }; return unsafe.String((*byte)(unsafe.Pointer(pointer)), length) }
func (playdateContext) NewBitmap(width, height int) (sdkPlaydate.Bitmap, error) {
	if err := sdkRuntime.ValidateBitmapSize(width, height); err != nil { return nil, err }
	handle := bridgeNewBitmap(int32(width), int32(height)); if handle == 0 { return nil, sdkPlaydate.ErrBitmapCreate }
	return sdkRuntime.NewOwnedBitmap(handle, bitmapDriver), nil
}
func (playdateContext) DrawBitmap(bitmap sdkPlaydate.Bitmap, x, y int) error {
	handle, err := sdkRuntime.BitmapHandle(bitmap); if err != nil { return err }; bridgeDrawBitmap(handle, int32(x), int32(y)); return nil
}
func (playdateContext) DrawScaledBitmap(bitmap sdkPlaydate.Bitmap, x, y int, scaleX, scaleY float32) error {
	if err := sdkRuntime.ValidateBitmapScale(scaleX, scaleY); err != nil { return err }
	handle, err := sdkRuntime.BitmapHandle(bitmap); if err != nil { return err }
	bridgeDrawScaledBitmapBits(handle, int32(x), int32(y), *(*uint32)(unsafe.Pointer(&scaleX)), *(*uint32)(unsafe.Pointer(&scaleY))); return nil
}
func drawPrimitive(kind, x1, y1, x2, y2, x3, y3, lineWidth int, startAngle, endAngle float32, paint sdkPlaydate.Paint) { solid, pattern, patterned := paint.Components(); flag := int32(0); if patterned { flag = 1 }; bridgeDrawPrimitive(int32(kind),int32(x1),int32(y1),int32(x2),int32(y2),int32(x3),int32(y3),int32(lineWidth),int32(solid),float32Bits(startAngle),float32Bits(endAngle),uintptr(unsafe.Pointer(&pattern[0])),flag) }
func (playdateContext) DrawLine(x1,y1,x2,y2,width int, paint sdkPlaydate.Paint) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,1,1,0,0); err!=nil{return err}; drawPrimitive(0,x1,y1,x2,y2,0,0,width,0,0,paint); return nil }
func (playdateContext) DrawRect(x,y,width,height int, paint sdkPlaydate.Paint) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,height,1,0,0); err!=nil{return err}; drawPrimitive(1,x,y,width,height,0,0,1,0,0,paint); return nil }
func (playdateContext) FillRect(x,y,width,height int, paint sdkPlaydate.Paint) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,height,1,0,0); err!=nil{return err}; drawPrimitive(2,x,y,width,height,0,0,1,0,0,paint); return nil }
func (playdateContext) DrawEllipse(x,y,width,height,lineWidth int,startAngle,endAngle float32,paint sdkPlaydate.Paint) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,height,lineWidth,startAngle,endAngle);err!=nil{return err};drawPrimitive(3,x,y,width,height,0,0,lineWidth,startAngle,endAngle,paint);return nil }
func (playdateContext) FillEllipse(x,y,width,height int,startAngle,endAngle float32,paint sdkPlaydate.Paint) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,height,1,startAngle,endAngle);err!=nil{return err};drawPrimitive(4,x,y,width,height,0,0,1,startAngle,endAngle,paint);return nil }
func (playdateContext) FillTriangle(x1,y1,x2,y2,x3,y3 int,paint sdkPlaydate.Paint) error { drawPrimitive(5,x1,y1,x2,y2,x3,y3,1,0,0,paint);return nil }
func (playdateContext) DrawTriangle(x1,y1,x2,y2,x3,y3,width int,paint sdkPlaydate.Paint) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,1,1,0,0);err!=nil{return err};drawPrimitive(6,x1,y1,x2,y2,x3,y3,width,0,0,paint);return nil }
func (playdateContext) SetClipRect(x,y,width,height int) error { if err:=sdkRuntime.ValidatePrimitiveGeometry(width,height,1,0,0);err!=nil{return err};bridgeSetClipRect(int32(x),int32(y),int32(width),int32(height));return nil }
func (playdateContext) ClearClipRect(){bridgeClearClipRect()}
func (playdateContext) SetDrawOffset(dx,dy int){bridgeSetDrawOffset(int32(dx),int32(dy))}
func (playdateContext) SetDrawMode(mode sdkPlaydate.DrawMode) error {if err:=sdkRuntime.ValidateDrawMode(mode);err!=nil{return err};bridgeSetDrawMode(int32(mode));return nil}

func float32Bits(value float32) uint32 { return *(*uint32)(unsafe.Pointer(&value)) }
var spriteDriver = sdkRuntime.SpriteDriver{
	SetBitmap: bridgeSpriteSetBitmap,
	MoveTo: func(sprite uintptr, x, y float32) { bridgeSpriteMoveToBits(sprite, float32Bits(x), float32Bits(y)) },
	MoveBy: func(sprite uintptr, dx, dy float32) { bridgeSpriteMoveByBits(sprite, float32Bits(dx), float32Bits(dy)) },
	SetVisible: func(sprite uintptr, visible bool) { value := int32(0); if visible { value = 1 }; bridgeSpriteSetVisible(sprite, value) },
	SetZIndex: func(sprite uintptr, z int) { bridgeSpriteSetZIndex(sprite, int32(z)) },
	SetCollideRect: func(sprite uintptr, rect sdkPlaydate.Rect) { bridgeSpriteSetCollideRectBits(sprite, float32Bits(rect.X), float32Bits(rect.Y), float32Bits(rect.Width), float32Bits(rect.Height)) },
	ClearCollideRect: bridgeSpriteClearCollideRect,
	SetTag: bridgeSpriteSetTag,
	MoveWithCollisions: moveWithCollisions,
	Add: bridgeSpriteAdd, Remove: bridgeSpriteRemove, Free: bridgeFreeSprite,
}
func moveWithCollisions(sprite uintptr, x, y float32) (float32, float32, []sdkRuntime.NativeCollision) {
	var actualX, actualY uint32; var count int32; list := bridgeSpriteMoveWithCollisionsBits(sprite, float32Bits(x), float32Bits(y), &actualX, &actualY, &count)
	result := make([]sdkRuntime.NativeCollision, int(count)); for index := range result { value := func(field int32) float32 { return float32FromBits(bridgeCollisionValueBits(list, int32(index), field)) }; result[index] = sdkRuntime.NativeCollision{Other: bridgeCollisionOther(list, int32(index)), ResponseType: sdkPlaydate.CollisionResponse(bridgeCollisionValueBits(list, int32(index), 14)), Overlaps: bridgeCollisionValueBits(list, int32(index), 15) != 0, Time: value(0), Move: sdkPlaydate.Point{X:value(1),Y:value(2)}, Normal:sdkPlaydate.Point{X:value(3),Y:value(4)}, Touch:sdkPlaydate.Point{X:value(5),Y:value(6)}, SpriteRect:sdkPlaydate.Rect{X:value(7),Y:value(8),Width:value(9),Height:value(10)}, OtherRect:sdkPlaydate.Rect{X:value(11),Y:value(12),Width:value(13),Height:value(16)}} }; bridgeFreeList(list); return float32FromBits(actualX), float32FromBits(actualY), result
}
func (playdateContext) NewSprite() (sdkPlaydate.Sprite, error) {
	handle := bridgeNewSprite(); if handle == 0 { return nil, sdkPlaydate.ErrSpriteCreate }
	return sdkRuntime.NewOwnedSprite(handle, spriteDriver), nil
}
func borrowedSpriteList(list uintptr, count int32) []sdkPlaydate.Sprite { handles := make([]uintptr, int(count)); for index := range handles { handles[index] = bridgeSpriteListItem(list, int32(index)) }; sprites := sdkRuntime.BorrowedSprites(handles, spriteDriver); bridgeFreeList(list); return sprites }
func (playdateContext) QuerySpritesAtPoint(x, y float32) []sdkPlaydate.Sprite { var count int32; list := bridgeQuerySpritesAtPointBits(float32Bits(x), float32Bits(y), &count); return borrowedSpriteList(list, count) }
func (playdateContext) QuerySpritesInRect(rect sdkPlaydate.Rect) []sdkPlaydate.Sprite { var count int32; list := bridgeQuerySpritesInRectBits(float32Bits(rect.X), float32Bits(rect.Y), float32Bits(rect.Width), float32Bits(rect.Height), &count); return borrowedSpriteList(list, count) }
func (playdateContext) QueryOverlappingSprites(sprite sdkPlaydate.Sprite) ([]sdkPlaydate.Sprite, error) { handle, err := sdkRuntime.SpriteHandle(sprite); if err != nil { return nil, err }; var count int32; list := bridgeOverlappingSprites(handle, &count); return borrowedSpriteList(list, count), nil }
func (playdateContext) UpdateAndDrawSprites() { bridgeUpdateAndDrawSprites() }

var soundEffectDriver = sdkRuntime.AudioDriver{
	Play: func(handle uintptr) bool { return bridgeSoundEffectPlay(handle) != 0 }, Stop: bridgeSoundEffectStop,
	SetVolume: func(handle uintptr, left, right float32) { bridgeSoundEffectSetVolumeBits(handle, float32Bits(left), float32Bits(right)) },
	Volume: func(handle uintptr) (float32, float32) { var left, right uint32; bridgeSoundEffectVolumeBits(handle, &left, &right); return float32FromBits(left), float32FromBits(right) },
	IsPlaying: func(handle uintptr) bool { return bridgeSoundEffectIsPlaying(handle) != 0 },
	Pause: func(handle uintptr, paused bool) { value := int32(0); if paused { value = 1 }; bridgeSoundEffectPause(handle, value) }, Free: bridgeFreeSoundEffect,
}
var filePlayerDriver = sdkRuntime.AudioDriver{
	Play: func(handle uintptr) bool { return bridgeFilePlayerPlay(handle) != 0 }, Stop: bridgeFilePlayerStop,
	SetVolume: func(handle uintptr, left, right float32) { bridgeFilePlayerSetVolumeBits(handle, float32Bits(left), float32Bits(right)) },
	Volume: func(handle uintptr) (float32, float32) { var left, right uint32; bridgeFilePlayerVolumeBits(handle, &left, &right); return float32FromBits(left), float32FromBits(right) },
	IsPlaying: func(handle uintptr) bool { return bridgeFilePlayerIsPlaying(handle) != 0 },
	Pause: func(handle uintptr, _ bool) { bridgeFilePlayerPause(handle) }, Free: bridgeFreeFilePlayer,
}
func (playdateContext) LoadSoundEffect(path string) (sdkPlaydate.SoundEffect, error) { terminated := path + "\x00"; handle := bridgeLoadSoundEffect(unsafe.StringData(terminated)); if handle == 0 { return nil, sdkPlaydate.AudioLoadError(path) }; return sdkRuntime.NewSoundEffect(handle, soundEffectDriver), nil }
func (playdateContext) LoadFilePlayer(path string) (sdkPlaydate.FilePlayer, error) { terminated := path + "\x00"; handle := bridgeLoadFilePlayer(unsafe.StringData(terminated)); if handle == 0 { return nil, sdkPlaydate.AudioLoadError(path) }; return sdkRuntime.NewFilePlayer(handle, filePlayerDriver), nil }

var gameContext playdateContext
var application = mustApplication()

func mustApplication() *sdkRuntime.Application {
	application, err := sdkRuntime.NewApplication(app.New(), gameContext, nil)
	if err != nil {
		panic(err)
	}
	return application
}

//export goEventHandler
func goEventHandler(_ uintptr, event int32, arg uint32) int32 {
	if err := application.Handle(sdkRuntime.Event(event), arg); err != nil {
		return -1
	}
	return 0
}

//export goUpdate
func goUpdate() int32 {
	refresh, err := application.Update(sdkRuntime.RawInput{
		Buttons: sdkPlaydate.Buttons(bridgeButtons()), CrankAngle: float32FromBits(bridgeCrankAngleBits()),
		CrankDelta: float32FromBits(bridgeCrankDeltaBits()), CrankDocked: bridgeCrankDocked() != 0,
		DeltaSeconds: float32FromBits(bridgeFrameDeltaBits()),
	})
	if err != nil {
		return 0
	}
	return int32(refresh)
}

func main() {}
`, modulePath+"/internal/features/runtime", modulePath+"/playdate", applicationImport)
}

type applicationInfo struct {
	ImportPath string
	Name       string
	Dir        string
	Module     *struct{ Path, Dir, GoVersion string }
}

func renderDeviceGoMod(module gomodule.Info, app applicationInfo) string {
	contents := gomodule.RenderProbe(module, "probe/device")
	if app.Module == nil || app.Module.Path == module.Path {
		return contents
	}
	return contents + fmt.Sprintf("\nrequire %s v0.0.0\nreplace %s => %s\n",
		app.Module.Path, app.Module.Path, strconv.Quote(filepath.ToSlash(app.Module.Dir)))
}

func inspectApplication(ctx context.Context, pattern string) (applicationInfo, error) {
	output, err := exec.CommandContext(ctx, "go", "list", "-json", pattern).CombinedOutput()
	if err != nil {
		return applicationInfo{}, commandError("inspect application package", err, output)
	}
	var app applicationInfo
	if err := json.Unmarshal(output, &app); err != nil {
		return applicationInfo{}, fmt.Errorf("inspect application package: decode go list output: %w", err)
	}
	return app, nil
}

const bootstrapSource = `#include "pd_api.h"

_Static_assert(sizeof(PDSystemEvent) <= 4, "PDSystemEvent must fit a 32-bit call slot");
_Static_assert(kEventMirrorEnded <= INT32_MAX, "PDSystemEvent values must fit int32_t");
_Static_assert(sizeof(uint32_t) == 4, "event argument must be 32-bit");
_Static_assert(sizeof(uintptr_t) == 4, "device pointers must be 32-bit");
_Static_assert(sizeof(int) == 4, "Playdate callback result must be 32-bit");
_Static_assert(sizeof(float) == 4, "Playdate float samples must be IEEE-754 binary32 slots");

extern void runtimeRun(void) __asm__("runtime.run");
extern int goEventHandler(PlaydateAPI*, PDSystemEvent, uint32_t);
extern int goUpdate(void);
void* runtimeAlloc(uintptr_t, void*) __asm__("runtime.alloc");

static int booted;
static PlaydateAPI* activePlaydate;
static int bridgeUpdate(void* userdata);

void* runtimeAlloc(uintptr_t size, void* layout)
{
    (void)layout;
    unsigned char* pointer = activePlaydate->system->realloc(NULL, size);
    if (pointer == NULL)
        return NULL;
    for (uintptr_t index = 0; index < size; ++index)
        pointer[index] = 0;
    return pointer;
}

int eventHandler(PlaydateAPI* playdate, PDSystemEvent event, uint32_t arg)
{
	int result;
    if (event == kEventInit && !booted) {
		activePlaydate = playdate;
        runtimeRun();
        booted = 1;
    }
	result = goEventHandler(playdate, event, arg);
	if (event == kEventInit && result == 0)
	{
		playdate->system->setUpdateCallback(bridgeUpdate, playdate);
		playdate->system->resetElapsedTime();
	}
	return result;
}

static int bridgeUpdate(void* userdata)
{
	(void)userdata;
	return goUpdate();
}

void bridgeClear(void)
{
	activePlaydate->graphics->clear(kColorWhite);
}

void bridgeDrawText(const char* text, uintptr_t length, int32_t x, int32_t y)
{
	activePlaydate->graphics->drawText(text, length, kUTF8Encoding, x, y);
}
uintptr_t bridgeLoadFont(const char* path, uintptr_t* error) { return (uintptr_t)activePlaydate->graphics->loadFont(path, (const char**)error); }
void bridgeSetFont(uintptr_t font) { activePlaydate->graphics->setFont((LCDFont*)font); }
int32_t bridgeTextWidth(uintptr_t font, const char* text, uintptr_t length) { return activePlaydate->graphics->getTextWidth((LCDFont*)font,text,length,kUTF8Encoding,0); }
int32_t bridgeFontHeight(uintptr_t font) { return activePlaydate->graphics->getFontHeight((LCDFont*)font); }
void bridgeFreeFont(uintptr_t font) { activePlaydate->system->realloc((void*)font,0); }

uint32_t bridgeCurrentTimeMilliseconds(void)
{
	return activePlaydate->system->getCurrentTimeMilliseconds();
}

uint32_t bridgeButtons(void) { PDButtons value = 0; activePlaydate->system->getButtonState(&value, NULL, NULL); return (uint32_t)value; }
static uint32_t bridgeFloatBits(float value) { union { float value; uint32_t bits; } conversion = { .value = value }; return conversion.bits; }
uint32_t bridgeCrankAngleBits(void) { return bridgeFloatBits(activePlaydate->system->getCrankAngle()); }
uint32_t bridgeCrankDeltaBits(void) { return bridgeFloatBits(activePlaydate->system->getCrankChange()); }
int32_t bridgeCrankDocked(void) { return activePlaydate->system->isCrankDocked(); }
uint32_t bridgeFrameDeltaBits(void) { float value = activePlaydate->system->getElapsedTime(); activePlaydate->system->resetElapsedTime(); return bridgeFloatBits(value); }
uintptr_t bridgeLoadBitmap(const char* path, const char** error) { return (uintptr_t)activePlaydate->graphics->loadBitmap(path, error); }
uintptr_t bridgeLoadBitmapTable(const char* path, const char** error) { return (uintptr_t)activePlaydate->graphics->loadBitmapTable(path, error); }
uintptr_t bridgeBitmapTableFrame(uintptr_t table, int32_t index) { return (uintptr_t)activePlaydate->graphics->getTableBitmap((LCDBitmapTable*)table, index); }
void bridgeFreeBitmapTable(uintptr_t table) { activePlaydate->graphics->freeBitmapTable((LCDBitmapTable*)table); }
uintptr_t bridgeNewBitmap(int32_t width, int32_t height) { return (uintptr_t)activePlaydate->graphics->newBitmap(width, height, kColorClear); }
void bridgeFreeBitmap(uintptr_t bitmap) { activePlaydate->graphics->freeBitmap((LCDBitmap*)bitmap); }
void bridgeBitmapSize(uintptr_t bitmap, int32_t* width, int32_t* height) { int nativeWidth, nativeHeight; activePlaydate->graphics->getBitmapData((LCDBitmap*)bitmap, &nativeWidth, &nativeHeight, NULL, NULL, NULL); *width = nativeWidth; *height = nativeHeight; }
static LCDColor bridgeBitmapColor(int32_t color) { return color == 1 ? kColorWhite : color == 2 ? kColorBlack : kColorClear; }
void bridgeFillBitmap(uintptr_t bitmap, int32_t color) { activePlaydate->graphics->clearBitmap((LCDBitmap*)bitmap, bridgeBitmapColor(color)); }
void bridgeDrawBitmap(uintptr_t bitmap, int32_t x, int32_t y) { activePlaydate->graphics->drawBitmap((LCDBitmap*)bitmap, x, y, kBitmapUnflipped); }
void bridgeDrawScaledBitmapBits(uintptr_t bitmap, int32_t x, int32_t y, uint32_t scaleX, uint32_t scaleY) { union { uint32_t bits; float value; } sx = { .bits = scaleX }, sy = { .bits = scaleY }; activePlaydate->graphics->drawScaledBitmap((LCDBitmap*)bitmap, x, y, sx.value, sy.value); }
static LCDColor bridgePrimitivePaint(int32_t solid, uintptr_t pattern, int32_t patterned) { if(patterned) return (LCDColor)pattern; return solid==1?kColorWhite:solid==2?kColorBlack:solid==3?kColorXOR:kColorClear; }
void bridgeDrawPrimitive(int32_t kind,int32_t x1,int32_t y1,int32_t x2,int32_t y2,int32_t x3,int32_t y3,int32_t lineWidth,int32_t solid,uint32_t startAngle,uint32_t endAngle,uintptr_t pattern,int32_t patterned) { union{uint32_t bits;float value;} start={.bits=startAngle},end={.bits=endAngle}; LCDColor color=bridgePrimitivePaint(solid,pattern,patterned); switch(kind){case 0:activePlaydate->graphics->drawLine(x1,y1,x2,y2,lineWidth,color);break;case 1:activePlaydate->graphics->drawRect(x1,y1,x2,y2,color);break;case 2:activePlaydate->graphics->fillRect(x1,y1,x2,y2,color);break;case 3:activePlaydate->graphics->drawEllipse(x1,y1,x2,y2,lineWidth,start.value,end.value,color);break;case 4:activePlaydate->graphics->fillEllipse(x1,y1,x2,y2,start.value,end.value,color);break;case 5:activePlaydate->graphics->fillTriangle(x1,y1,x2,y2,x3,y3,color);break;case 6:activePlaydate->graphics->drawLine(x1,y1,x2,y2,lineWidth,color);activePlaydate->graphics->drawLine(x2,y2,x3,y3,lineWidth,color);activePlaydate->graphics->drawLine(x3,y3,x1,y1,lineWidth,color);break;} }
void bridgeSetClipRect(int32_t x,int32_t y,int32_t width,int32_t height){activePlaydate->graphics->setClipRect(x,y,width,height);}
void bridgeClearClipRect(void){activePlaydate->graphics->clearClipRect();}
void bridgeSetDrawOffset(int32_t dx,int32_t dy){activePlaydate->graphics->setDrawOffset(dx,dy);}
void bridgeSetDrawMode(int32_t mode){activePlaydate->graphics->setDrawMode((LCDBitmapDrawMode)mode);}
uintptr_t bridgeNewSprite(void) { return (uintptr_t)activePlaydate->sprite->newSprite(); }
void bridgeFreeSprite(uintptr_t sprite) { activePlaydate->sprite->freeSprite((LCDSprite*)sprite); }
void bridgeSpriteSetBitmap(uintptr_t sprite, uintptr_t bitmap) { activePlaydate->sprite->setImage((LCDSprite*)sprite, (LCDBitmap*)bitmap, kBitmapUnflipped); }
void bridgeSpriteMoveToBits(uintptr_t sprite, uint32_t x, uint32_t y) { union { uint32_t bits; float value; } px = { .bits = x }, py = { .bits = y }; activePlaydate->sprite->moveTo((LCDSprite*)sprite, px.value, py.value); }
void bridgeSpriteMoveByBits(uintptr_t sprite, uint32_t dx, uint32_t dy) { union { uint32_t bits; float value; } px = { .bits = dx }, py = { .bits = dy }; activePlaydate->sprite->moveBy((LCDSprite*)sprite, px.value, py.value); }
void bridgeSpriteSetVisible(uintptr_t sprite, int32_t visible) { activePlaydate->sprite->setVisible((LCDSprite*)sprite, visible); }
void bridgeSpriteSetZIndex(uintptr_t sprite, int32_t z) { activePlaydate->sprite->setZIndex((LCDSprite*)sprite, z); }
void bridgeSpriteSetCollideRectBits(uintptr_t sprite, uint32_t x, uint32_t y, uint32_t width, uint32_t height) { union { uint32_t bits; float value; } px={.bits=x}, py={.bits=y}, pw={.bits=width}, ph={.bits=height}; activePlaydate->sprite->setCollideRect((LCDSprite*)sprite, (PDRect){px.value,py.value,pw.value,ph.value}); }
void bridgeSpriteClearCollideRect(uintptr_t sprite) { activePlaydate->sprite->clearCollideRect((LCDSprite*)sprite); }
void bridgeSpriteSetTag(uintptr_t sprite, uint8_t tag) { activePlaydate->sprite->setTag((LCDSprite*)sprite, tag); }
uintptr_t bridgeSpriteMoveWithCollisionsBits(uintptr_t sprite, uint32_t x, uint32_t y, uint32_t* actualX, uint32_t* actualY, int32_t* count) { union { uint32_t bits; float value; } px={.bits=x}, py={.bits=y}, ax, ay; SpriteCollisionInfo* result=activePlaydate->sprite->moveWithCollisions((LCDSprite*)sprite,px.value,py.value,&ax.value,&ay.value,(int*)count); *actualX=ax.bits; *actualY=ay.bits; return (uintptr_t)result; }
uint32_t bridgeCollisionValueBits(uintptr_t collisions, int32_t index, int32_t field) { SpriteCollisionInfo* c=&((SpriteCollisionInfo*)collisions)[index]; union { float value; uint32_t bits; } v; switch(field){case 0:v.value=c->ti;break;case 1:v.value=c->move.x;break;case 2:v.value=c->move.y;break;case 3:v.value=c->normal.x;break;case 4:v.value=c->normal.y;break;case 5:v.value=c->touch.x;break;case 6:v.value=c->touch.y;break;case 7:v.value=c->spriteRect.x;break;case 8:v.value=c->spriteRect.y;break;case 9:v.value=c->spriteRect.width;break;case 10:v.value=c->spriteRect.height;break;case 11:v.value=c->otherRect.x;break;case 12:v.value=c->otherRect.y;break;case 13:v.value=c->otherRect.width;break;case 16:v.value=c->otherRect.height;break;case 14:return (uint32_t)c->responseType;case 15:return (uint32_t)c->overlaps;default:return 0;} return v.bits; }
uintptr_t bridgeCollisionOther(uintptr_t collisions, int32_t index) { return (uintptr_t)((SpriteCollisionInfo*)collisions)[index].other; }
void bridgeFreeList(uintptr_t list) { if(list) activePlaydate->system->realloc((void*)list,0); }
uintptr_t bridgeQuerySpritesAtPointBits(uint32_t x,uint32_t y,int32_t* count) { union {uint32_t bits;float value;} px={.bits=x},py={.bits=y}; return (uintptr_t)activePlaydate->sprite->querySpritesAtPoint(px.value,py.value,(int*)count); }
uintptr_t bridgeQuerySpritesInRectBits(uint32_t x,uint32_t y,uint32_t width,uint32_t height,int32_t* count) { union {uint32_t bits;float value;} px={.bits=x},py={.bits=y},pw={.bits=width},ph={.bits=height}; return (uintptr_t)activePlaydate->sprite->querySpritesInRect(px.value,py.value,pw.value,ph.value,(int*)count); }
uintptr_t bridgeSpriteListItem(uintptr_t list,int32_t index) { return (uintptr_t)((LCDSprite**)list)[index]; }
uintptr_t bridgeOverlappingSprites(uintptr_t sprite,int32_t* count) { return (uintptr_t)activePlaydate->sprite->overlappingSprites((LCDSprite*)sprite,(int*)count); }
void bridgeSpriteAdd(uintptr_t sprite) { activePlaydate->sprite->addSprite((LCDSprite*)sprite); }
void bridgeSpriteRemove(uintptr_t sprite) { activePlaydate->sprite->removeSprite((LCDSprite*)sprite); }
void bridgeUpdateAndDrawSprites(void) { activePlaydate->sprite->updateAndDrawSprites(); }
typedef struct { AudioSample* sample; SamplePlayer* player; } BridgeSoundEffect;
uintptr_t bridgeLoadSoundEffect(const char* path) { AudioSample* sample=activePlaydate->sound->sample->load(path); if(!sample)return 0; SamplePlayer* player=activePlaydate->sound->sampleplayer->newPlayer(); if(!player){activePlaydate->sound->sample->freeSample(sample);return 0;} BridgeSoundEffect* effect=activePlaydate->system->realloc(NULL,sizeof(BridgeSoundEffect)); if(!effect){activePlaydate->sound->sampleplayer->freePlayer(player);activePlaydate->sound->sample->freeSample(sample);return 0;} effect->sample=sample;effect->player=player;activePlaydate->sound->sampleplayer->setSample(player,sample);return(uintptr_t)effect; }
static BridgeSoundEffect* bridgeEffect(uintptr_t effect){return(BridgeSoundEffect*)effect;}
int32_t bridgeSoundEffectPlay(uintptr_t effect){return activePlaydate->sound->sampleplayer->play(bridgeEffect(effect)->player,1,1.0f);}
void bridgeSoundEffectStop(uintptr_t effect){activePlaydate->sound->sampleplayer->stop(bridgeEffect(effect)->player);}
void bridgeSoundEffectSetVolumeBits(uintptr_t effect,uint32_t left,uint32_t right){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->sampleplayer->setVolume(bridgeEffect(effect)->player,l.value,r.value);}
void bridgeSoundEffectVolumeBits(uintptr_t effect,uint32_t* left,uint32_t* right){union{float value;uint32_t bits;}l,r;activePlaydate->sound->sampleplayer->getVolume(bridgeEffect(effect)->player,&l.value,&r.value);*left=l.bits;*right=r.bits;}
int32_t bridgeSoundEffectIsPlaying(uintptr_t effect){return activePlaydate->sound->sampleplayer->isPlaying(bridgeEffect(effect)->player);}
void bridgeSoundEffectPause(uintptr_t effect,int32_t paused){activePlaydate->sound->sampleplayer->setPaused(bridgeEffect(effect)->player,paused);}
void bridgeFreeSoundEffect(uintptr_t effect){BridgeSoundEffect* value=bridgeEffect(effect);activePlaydate->sound->sampleplayer->freePlayer(value->player);activePlaydate->sound->sample->freeSample(value->sample);activePlaydate->system->realloc(value,0);}
uintptr_t bridgeLoadFilePlayer(const char* path){FilePlayer* player=activePlaydate->sound->fileplayer->newPlayer();if(!player)return 0;if(!activePlaydate->sound->fileplayer->loadIntoPlayer(player,path)){activePlaydate->sound->fileplayer->freePlayer(player);return 0;}return(uintptr_t)player;}
int32_t bridgeFilePlayerPlay(uintptr_t player){return activePlaydate->sound->fileplayer->play((FilePlayer*)player,1);}
void bridgeFilePlayerStop(uintptr_t player){activePlaydate->sound->fileplayer->stop((FilePlayer*)player);}
void bridgeFilePlayerSetVolumeBits(uintptr_t player,uint32_t left,uint32_t right){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->fileplayer->setVolume((FilePlayer*)player,l.value,r.value);}
void bridgeFilePlayerVolumeBits(uintptr_t player,uint32_t* left,uint32_t* right){union{float value;uint32_t bits;}l,r;activePlaydate->sound->fileplayer->getVolume((FilePlayer*)player,&l.value,&r.value);*left=l.bits;*right=r.bits;}
int32_t bridgeFilePlayerIsPlaying(uintptr_t player){return activePlaydate->sound->fileplayer->isPlaying((FilePlayer*)player);}
void bridgeFilePlayerPause(uintptr_t player){activePlaydate->sound->fileplayer->pause((FilePlayer*)player);}
void bridgeFreeFilePlayer(uintptr_t player){activePlaydate->sound->fileplayer->freePlayer((FilePlayer*)player);}
`

const conservativeBootstrapSource = `#include "pd_api.h"

_Static_assert(sizeof(PDSystemEvent) <= 4, "PDSystemEvent must fit a 32-bit call slot");
_Static_assert(kEventMirrorEnded <= INT32_MAX, "PDSystemEvent values must fit int32_t");
_Static_assert(sizeof(uint32_t) == 4, "event argument must be 32-bit");
_Static_assert(sizeof(uintptr_t) == 4, "device pointers must be 32-bit");
_Static_assert(sizeof(int) == 4, "Playdate callback result must be 32-bit");
_Static_assert(sizeof(float) == 4, "Playdate float samples must be IEEE-754 binary32 slots");

extern void runtimeRun(void) __asm__("runtime.run");
extern uintptr_t runtimeStackTop __asm__("runtime.stackTop");
extern void* runtimeSCB __asm__("playdateRuntimeSCB");
extern int goEventHandler(PlaydateAPI*, PDSystemEvent, uint32_t);
extern int goUpdate(void);

static int booted;
static PlaydateAPI* activePlaydate;
static uint32_t runtimeSCBShadow[2];
__attribute__((section(".bss.playdate_runtime_heap"), aligned(16), used))
unsigned char playdateRuntimeHeap[256 * 1024];
static int bridgeUpdate(void* userdata);

static void prepareRuntimeBoundary(void)
{
	register uintptr_t stackPointer __asm__("sp");
	runtimeStackTop = stackPointer;
}

int eventHandler(PlaydateAPI* playdate, PDSystemEvent event, uint32_t arg)
{
	int result;
    if (event == kEventInit && !booted) {
		activePlaydate = playdate;
		runtimeSCB = runtimeSCBShadow;
		prepareRuntimeBoundary();
        runtimeRun();
        booted = 1;
    }
	result = goEventHandler(playdate, event, arg);
	if (event == kEventInit && result == 0)
	{
		playdate->system->setUpdateCallback(bridgeUpdate, playdate);
		playdate->system->resetElapsedTime();
	}
	return result;
}

static int bridgeUpdate(void* userdata)
{
	(void)userdata;
	prepareRuntimeBoundary();
	return goUpdate();
}

void bridgeClear(void)
{
	activePlaydate->graphics->clear(kColorWhite);
}

void bridgeDrawText(const char* text, uintptr_t length, int32_t x, int32_t y)
{
	activePlaydate->graphics->drawText(text, length, kUTF8Encoding, x, y);
}
uintptr_t bridgeLoadFont(const char* path, uintptr_t* error) { return (uintptr_t)activePlaydate->graphics->loadFont(path, (const char**)error); }
void bridgeSetFont(uintptr_t font) { activePlaydate->graphics->setFont((LCDFont*)font); }
int32_t bridgeTextWidth(uintptr_t font, const char* text, uintptr_t length) { return activePlaydate->graphics->getTextWidth((LCDFont*)font,text,length,kUTF8Encoding,0); }
int32_t bridgeFontHeight(uintptr_t font) { return activePlaydate->graphics->getFontHeight((LCDFont*)font); }
void bridgeFreeFont(uintptr_t font) { activePlaydate->system->realloc((void*)font,0); }

uint32_t bridgeCurrentTimeMilliseconds(void)
{
	return activePlaydate->system->getCurrentTimeMilliseconds();
}

uint32_t bridgeButtons(void) { PDButtons value = 0; activePlaydate->system->getButtonState(&value, NULL, NULL); return (uint32_t)value; }
static uint32_t bridgeFloatBits(float value) { union { float value; uint32_t bits; } conversion = { .value = value }; return conversion.bits; }
uint32_t bridgeCrankAngleBits(void) { return bridgeFloatBits(activePlaydate->system->getCrankAngle()); }
uint32_t bridgeCrankDeltaBits(void) { return bridgeFloatBits(activePlaydate->system->getCrankChange()); }
int32_t bridgeCrankDocked(void) { return activePlaydate->system->isCrankDocked(); }
uint32_t bridgeFrameDeltaBits(void) { float value = activePlaydate->system->getElapsedTime(); activePlaydate->system->resetElapsedTime(); return bridgeFloatBits(value); }
uintptr_t bridgeLoadBitmap(const char* path, const char** error) { return (uintptr_t)activePlaydate->graphics->loadBitmap(path, error); }
uintptr_t bridgeLoadBitmapTable(const char* path, const char** error) { return (uintptr_t)activePlaydate->graphics->loadBitmapTable(path, error); }
uintptr_t bridgeBitmapTableFrame(uintptr_t table, int32_t index) { return (uintptr_t)activePlaydate->graphics->getTableBitmap((LCDBitmapTable*)table, index); }
void bridgeFreeBitmapTable(uintptr_t table) { activePlaydate->graphics->freeBitmapTable((LCDBitmapTable*)table); }
uintptr_t bridgeNewBitmap(int32_t width, int32_t height) { return (uintptr_t)activePlaydate->graphics->newBitmap(width, height, kColorClear); }
void bridgeFreeBitmap(uintptr_t bitmap) { activePlaydate->graphics->freeBitmap((LCDBitmap*)bitmap); }
void bridgeBitmapSize(uintptr_t bitmap, int32_t* width, int32_t* height) { int nativeWidth, nativeHeight; activePlaydate->graphics->getBitmapData((LCDBitmap*)bitmap, &nativeWidth, &nativeHeight, NULL, NULL, NULL); *width = nativeWidth; *height = nativeHeight; }
static LCDColor bridgeBitmapColor(int32_t color) { return color == 1 ? kColorWhite : color == 2 ? kColorBlack : kColorClear; }
void bridgeFillBitmap(uintptr_t bitmap, int32_t color) { activePlaydate->graphics->clearBitmap((LCDBitmap*)bitmap, bridgeBitmapColor(color)); }
void bridgeDrawBitmap(uintptr_t bitmap, int32_t x, int32_t y) { activePlaydate->graphics->drawBitmap((LCDBitmap*)bitmap, x, y, kBitmapUnflipped); }
void bridgeDrawScaledBitmapBits(uintptr_t bitmap, int32_t x, int32_t y, uint32_t scaleX, uint32_t scaleY) { union { uint32_t bits; float value; } sx = { .bits = scaleX }, sy = { .bits = scaleY }; activePlaydate->graphics->drawScaledBitmap((LCDBitmap*)bitmap, x, y, sx.value, sy.value); }
static LCDColor bridgePrimitivePaint(int32_t solid, uintptr_t pattern, int32_t patterned) { if(patterned) return (LCDColor)pattern; return solid==1?kColorWhite:solid==2?kColorBlack:solid==3?kColorXOR:kColorClear; }
void bridgeDrawPrimitive(int32_t kind,int32_t x1,int32_t y1,int32_t x2,int32_t y2,int32_t x3,int32_t y3,int32_t lineWidth,int32_t solid,uint32_t startAngle,uint32_t endAngle,uintptr_t pattern,int32_t patterned) { union{uint32_t bits;float value;} start={.bits=startAngle},end={.bits=endAngle}; LCDColor color=bridgePrimitivePaint(solid,pattern,patterned); switch(kind){case 0:activePlaydate->graphics->drawLine(x1,y1,x2,y2,lineWidth,color);break;case 1:activePlaydate->graphics->drawRect(x1,y1,x2,y2,color);break;case 2:activePlaydate->graphics->fillRect(x1,y1,x2,y2,color);break;case 3:activePlaydate->graphics->drawEllipse(x1,y1,x2,y2,lineWidth,start.value,end.value,color);break;case 4:activePlaydate->graphics->fillEllipse(x1,y1,x2,y2,start.value,end.value,color);break;case 5:activePlaydate->graphics->fillTriangle(x1,y1,x2,y2,x3,y3,color);break;case 6:activePlaydate->graphics->drawLine(x1,y1,x2,y2,lineWidth,color);activePlaydate->graphics->drawLine(x2,y2,x3,y3,lineWidth,color);activePlaydate->graphics->drawLine(x3,y3,x1,y1,lineWidth,color);break;} }
void bridgeSetClipRect(int32_t x,int32_t y,int32_t width,int32_t height){activePlaydate->graphics->setClipRect(x,y,width,height);}
void bridgeClearClipRect(void){activePlaydate->graphics->clearClipRect();}
void bridgeSetDrawOffset(int32_t dx,int32_t dy){activePlaydate->graphics->setDrawOffset(dx,dy);}
void bridgeSetDrawMode(int32_t mode){activePlaydate->graphics->setDrawMode((LCDBitmapDrawMode)mode);}
uintptr_t bridgeNewSprite(void) { return (uintptr_t)activePlaydate->sprite->newSprite(); }
void bridgeFreeSprite(uintptr_t sprite) { activePlaydate->sprite->freeSprite((LCDSprite*)sprite); }
void bridgeSpriteSetBitmap(uintptr_t sprite, uintptr_t bitmap) { activePlaydate->sprite->setImage((LCDSprite*)sprite, (LCDBitmap*)bitmap, kBitmapUnflipped); }
void bridgeSpriteMoveToBits(uintptr_t sprite, uint32_t x, uint32_t y) { union { uint32_t bits; float value; } px = { .bits = x }, py = { .bits = y }; activePlaydate->sprite->moveTo((LCDSprite*)sprite, px.value, py.value); }
void bridgeSpriteMoveByBits(uintptr_t sprite, uint32_t dx, uint32_t dy) { union { uint32_t bits; float value; } px = { .bits = dx }, py = { .bits = dy }; activePlaydate->sprite->moveBy((LCDSprite*)sprite, px.value, py.value); }
void bridgeSpriteSetVisible(uintptr_t sprite, int32_t visible) { activePlaydate->sprite->setVisible((LCDSprite*)sprite, visible); }
void bridgeSpriteSetZIndex(uintptr_t sprite, int32_t z) { activePlaydate->sprite->setZIndex((LCDSprite*)sprite, z); }
void bridgeSpriteSetCollideRectBits(uintptr_t sprite, uint32_t x, uint32_t y, uint32_t width, uint32_t height) { union { uint32_t bits; float value; } px={.bits=x}, py={.bits=y}, pw={.bits=width}, ph={.bits=height}; activePlaydate->sprite->setCollideRect((LCDSprite*)sprite, (PDRect){px.value,py.value,pw.value,ph.value}); }
void bridgeSpriteClearCollideRect(uintptr_t sprite) { activePlaydate->sprite->clearCollideRect((LCDSprite*)sprite); }
void bridgeSpriteSetTag(uintptr_t sprite, uint8_t tag) { activePlaydate->sprite->setTag((LCDSprite*)sprite, tag); }
uintptr_t bridgeSpriteMoveWithCollisionsBits(uintptr_t sprite, uint32_t x, uint32_t y, uint32_t* actualX, uint32_t* actualY, int32_t* count) { union { uint32_t bits; float value; } px={.bits=x}, py={.bits=y}, ax, ay; SpriteCollisionInfo* result=activePlaydate->sprite->moveWithCollisions((LCDSprite*)sprite,px.value,py.value,&ax.value,&ay.value,(int*)count); *actualX=ax.bits; *actualY=ay.bits; return (uintptr_t)result; }
uint32_t bridgeCollisionValueBits(uintptr_t collisions, int32_t index, int32_t field) { SpriteCollisionInfo* c=&((SpriteCollisionInfo*)collisions)[index]; union { float value; uint32_t bits; } v; switch(field){case 0:v.value=c->ti;break;case 1:v.value=c->move.x;break;case 2:v.value=c->move.y;break;case 3:v.value=c->normal.x;break;case 4:v.value=c->normal.y;break;case 5:v.value=c->touch.x;break;case 6:v.value=c->touch.y;break;case 7:v.value=c->spriteRect.x;break;case 8:v.value=c->spriteRect.y;break;case 9:v.value=c->spriteRect.width;break;case 10:v.value=c->spriteRect.height;break;case 11:v.value=c->otherRect.x;break;case 12:v.value=c->otherRect.y;break;case 13:v.value=c->otherRect.width;break;case 16:v.value=c->otherRect.height;break;case 14:return (uint32_t)c->responseType;case 15:return (uint32_t)c->overlaps;default:return 0;} return v.bits; }
uintptr_t bridgeCollisionOther(uintptr_t collisions, int32_t index) { return (uintptr_t)((SpriteCollisionInfo*)collisions)[index].other; }
void bridgeFreeList(uintptr_t list) { if(list) activePlaydate->system->realloc((void*)list,0); }
uintptr_t bridgeQuerySpritesAtPointBits(uint32_t x,uint32_t y,int32_t* count) { union {uint32_t bits;float value;} px={.bits=x},py={.bits=y}; return (uintptr_t)activePlaydate->sprite->querySpritesAtPoint(px.value,py.value,(int*)count); }
uintptr_t bridgeQuerySpritesInRectBits(uint32_t x,uint32_t y,uint32_t width,uint32_t height,int32_t* count) { union {uint32_t bits;float value;} px={.bits=x},py={.bits=y},pw={.bits=width},ph={.bits=height}; return (uintptr_t)activePlaydate->sprite->querySpritesInRect(px.value,py.value,pw.value,ph.value,(int*)count); }
uintptr_t bridgeSpriteListItem(uintptr_t list,int32_t index) { return (uintptr_t)((LCDSprite**)list)[index]; }
uintptr_t bridgeOverlappingSprites(uintptr_t sprite,int32_t* count) { return (uintptr_t)activePlaydate->sprite->overlappingSprites((LCDSprite*)sprite,(int*)count); }
void bridgeSpriteAdd(uintptr_t sprite) { activePlaydate->sprite->addSprite((LCDSprite*)sprite); }
void bridgeSpriteRemove(uintptr_t sprite) { activePlaydate->sprite->removeSprite((LCDSprite*)sprite); }
void bridgeUpdateAndDrawSprites(void) { activePlaydate->sprite->updateAndDrawSprites(); }
typedef struct { AudioSample* sample; SamplePlayer* player; } BridgeSoundEffect;
uintptr_t bridgeLoadSoundEffect(const char* path) { AudioSample* sample=activePlaydate->sound->sample->load(path); if(!sample)return 0; SamplePlayer* player=activePlaydate->sound->sampleplayer->newPlayer(); if(!player){activePlaydate->sound->sample->freeSample(sample);return 0;} BridgeSoundEffect* effect=activePlaydate->system->realloc(NULL,sizeof(BridgeSoundEffect)); if(!effect){activePlaydate->sound->sampleplayer->freePlayer(player);activePlaydate->sound->sample->freeSample(sample);return 0;} effect->sample=sample;effect->player=player;activePlaydate->sound->sampleplayer->setSample(player,sample);return(uintptr_t)effect; }
static BridgeSoundEffect* bridgeEffect(uintptr_t effect){return(BridgeSoundEffect*)effect;}
int32_t bridgeSoundEffectPlay(uintptr_t effect){return activePlaydate->sound->sampleplayer->play(bridgeEffect(effect)->player,1,1.0f);}
void bridgeSoundEffectStop(uintptr_t effect){activePlaydate->sound->sampleplayer->stop(bridgeEffect(effect)->player);}
void bridgeSoundEffectSetVolumeBits(uintptr_t effect,uint32_t left,uint32_t right){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->sampleplayer->setVolume(bridgeEffect(effect)->player,l.value,r.value);}
void bridgeSoundEffectVolumeBits(uintptr_t effect,uint32_t* left,uint32_t* right){union{float value;uint32_t bits;}l,r;activePlaydate->sound->sampleplayer->getVolume(bridgeEffect(effect)->player,&l.value,&r.value);*left=l.bits;*right=r.bits;}
int32_t bridgeSoundEffectIsPlaying(uintptr_t effect){return activePlaydate->sound->sampleplayer->isPlaying(bridgeEffect(effect)->player);}
void bridgeSoundEffectPause(uintptr_t effect,int32_t paused){activePlaydate->sound->sampleplayer->setPaused(bridgeEffect(effect)->player,paused);}
void bridgeFreeSoundEffect(uintptr_t effect){BridgeSoundEffect* value=bridgeEffect(effect);activePlaydate->sound->sampleplayer->freePlayer(value->player);activePlaydate->sound->sample->freeSample(value->sample);activePlaydate->system->realloc(value,0);}
uintptr_t bridgeLoadFilePlayer(const char* path){FilePlayer* player=activePlaydate->sound->fileplayer->newPlayer();if(!player)return 0;if(!activePlaydate->sound->fileplayer->loadIntoPlayer(player,path)){activePlaydate->sound->fileplayer->freePlayer(player);return 0;}return(uintptr_t)player;}
int32_t bridgeFilePlayerPlay(uintptr_t player){return activePlaydate->sound->fileplayer->play((FilePlayer*)player,1);}
void bridgeFilePlayerStop(uintptr_t player){activePlaydate->sound->fileplayer->stop((FilePlayer*)player);}
void bridgeFilePlayerSetVolumeBits(uintptr_t player,uint32_t left,uint32_t right){union{uint32_t bits;float value;}l={.bits=left},r={.bits=right};activePlaydate->sound->fileplayer->setVolume((FilePlayer*)player,l.value,r.value);}
void bridgeFilePlayerVolumeBits(uintptr_t player,uint32_t* left,uint32_t* right){union{float value;uint32_t bits;}l,r;activePlaydate->sound->fileplayer->getVolume((FilePlayer*)player,&l.value,&r.value);*left=l.bits;*right=r.bits;}
int32_t bridgeFilePlayerIsPlaying(uintptr_t player){return activePlaydate->sound->fileplayer->isPlaying((FilePlayer*)player);}
void bridgeFilePlayerPause(uintptr_t player){activePlaydate->sound->fileplayer->pause((FilePlayer*)player);}
void bridgeFreeFilePlayer(uintptr_t player){activePlaydate->sound->fileplayer->freePlayer((FilePlayer*)player);}
`

const targetSource = `{
  "inherits": ["cortex-m7"],
  "llvm-target": "thumbv7em-unknown-unknown-eabihf",
  "relocation-model": "pic",
  "build-tags": ["qemu"],
  "serial": "none",
  "cflags": ["-mfloat-abi=hard", "-mfpu=fpv5-sp-d16"]
}
`

const adapterSource = `.syntax unified
.thumb

.section .text.tinygo_scanCurrentStack,"ax",%progbits
.global tinygo_scanCurrentStack
.type tinygo_scanCurrentStack, %function
.thumb_func
tinygo_scanCurrentStack:
    push {r4-r11, lr}
    mov r0, sp
    bl tinygo_scanstack
    add sp, #32
    pop {pc}
.size tinygo_scanCurrentStack, .-tinygo_scanCurrentStack

.section .text.DisableInterrupts,"ax",%progbits
.global DisableInterrupts
.type DisableInterrupts, %function
.thumb_func
DisableInterrupts:
    mrs r0, primask
    cpsid i
    bx lr

.section .text.EnableInterrupts,"ax",%progbits
.global EnableInterrupts
.type EnableInterrupts, %function
.thumb_func
EnableInterrupts:
    msr primask, r0
    bx lr

.section .text.SemihostingCall,"ax",%progbits
.global SemihostingCall
.type SemihostingCall, %function
.thumb_func
SemihostingCall:
    bx lr

.section .text._exit,"ax",%progbits
.global _exit
.type _exit, %function
.thumb_func
_exit:
    b _exit

.section .text._kill,"ax",%progbits
.global _kill
.type _kill, %function
.thumb_func
_kill:
    movs r0, #1
    negs r0, r0
    bx lr

.section .text._getpid,"ax",%progbits
.global _getpid
.type _getpid, %function
.thumb_func
_getpid:
    movs r0, #1
    bx lr
`
