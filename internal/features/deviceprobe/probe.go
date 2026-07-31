// Package deviceprobe verifies TinyGo ARM linking and Playdate packaging.
package deviceprobe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Djunichi/gopdsdk/internal/shared/gomodule"
)

// Result records the verified device toolchain stage.
type Result struct {
	TinyGo  string
	GCC     string
	Format  string
	Export  string
	Package string
	Deploy  string
	Run     string
	Pending string
}

// Config identifies the official Playdate SDK used for the link stage.
type Config struct {
	SDKPath      string
	Install      bool
	Run          bool
	ArtifactsDir string
}

// Probe compiles and links a structural Playdate device ELF.
func Probe(ctx context.Context, config Config) (Result, error) {
	if config.SDKPath == "" {
		return Result{}, fmt.Errorf("Playdate SDK path is required")
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
	module, err := gomodule.Locate(ctx)
	if err != nil {
		return Result{}, err
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
		{name: "go.mod", contents: gomodule.RenderProbe(module, "probe/device")},
		{name: "main.go", contents: renderProbeSource(module.Path)},
		{name: "playdate.json", contents: targetSource},
		{name: "adapter.S", contents: adapterSource},
	} {
		if err := os.WriteFile(filepath.Join(workDir, file.name), []byte(file.contents), 0o644); err != nil {
			return Result{}, fmt.Errorf("write %s: %w", file.name, err)
		}
	}
	objectPath := filepath.Join(workDir, "probe.o")
	if err := runCommand(ctx, workDir, "compile TinyGo Playdate object", tinygo,
		"build", "-target", filepath.Join(workDir, "playdate.json"), "-scheduler", "none", "-gc", "none", "-panic", "trap", "-opt", "0", "-o", objectPath, "."); err != nil {
		return Result{}, err
	}
	if err := runCommand(ctx, workDir, "expose TinyGo runtime bootstrap symbols", objcopy,
		"--globalize-symbol=runtime.run", objectPath); err != nil {
		return Result{}, err
	}
	bootstrapSourcePath := filepath.Join(workDir, "bootstrap.c")
	if err := os.WriteFile(bootstrapSourcePath, []byte(bootstrapSource), 0o644); err != nil {
		return Result{}, fmt.Errorf("write bootstrap.c: %w", err)
	}
	setupObject := filepath.Join(workDir, "setup.o")
	adapterObject := filepath.Join(workDir, "adapter.o")
	bootstrapObject := filepath.Join(workDir, "bootstrap.o")
	compileFlags := []string{"-c", "-mthumb", "-mcpu=cortex-m7", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16", "-O2", "-ffunction-sections", "-fdata-sections"}
	setupArgs := append(append([]string{}, compileFlags...), "-DTARGET_PLAYDATE=1", "-DTARGET_EXTENSION=1", "-I", filepath.Join(sdkPath, "C_API"), setupSource, "-o", setupObject)
	if err := runCommand(ctx, workDir, "compile official Playdate setup", gcc, setupArgs...); err != nil {
		return Result{}, err
	}
	adapterArgs := append(append([]string{}, compileFlags...), filepath.Join(workDir, "adapter.S"), "-o", adapterObject)
	if err := runCommand(ctx, workDir, "compile TinyGo runtime adapter", gcc, adapterArgs...); err != nil {
		return Result{}, err
	}
	bootstrapArgs := append(append([]string{}, compileFlags...), "-DTARGET_PLAYDATE=1", "-DTARGET_EXTENSION=1", "-I", filepath.Join(sdkPath, "C_API"), bootstrapSourcePath, "-o", bootstrapObject)
	if err := runCommand(ctx, workDir, "compile TinyGo runtime bootstrap", gcc, bootstrapArgs...); err != nil {
		return Result{}, err
	}
	elfPath := filepath.Join(workDir, "pdex.elf")
	if err := runCommand(ctx, workDir, "link Playdate device ELF", gcc,
		objectPath, setupObject, adapterObject, bootstrapObject,
		"-nostartfiles", "-mthumb", "-mcpu=cortex-m7", "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16",
		"-T", linkerScript, linkerFlags, "-o", elfPath); err != nil {
		return Result{}, err
	}
	command := exec.CommandContext(ctx, readelf, "-h", "-A", "-s", "-r", elfPath)
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
	for _, forbidden := range []string{"R_ARM_THM_MOVW_AB", "R_ARM_THM_MOVT_AB"} {
		if strings.Contains(inspection, forbidden) {
			return Result{}, fmt.Errorf("inspect ARM object: unsupported Playdate relocation %s remains", forbidden)
		}
	}
	undefined := exec.CommandContext(ctx, nm, "-u", elfPath)
	undefinedOutput, runErr := undefined.CombinedOutput()
	if runErr != nil {
		return Result{}, commandError("inspect unresolved ELF symbols", runErr, undefinedOutput)
	}
	if unresolved := strongUndefinedSymbols(string(undefinedOutput)); len(unresolved) != 0 {
		return Result{}, fmt.Errorf("inspect unresolved ELF symbols: %s", strings.Join(unresolved, ", "))
	}
	symbols := exec.CommandContext(ctx, nm, elfPath)
	symbolOutput, runErr := symbols.CombinedOutput()
	if runErr != nil {
		return Result{}, commandError("inspect linked ELF symbols", runErr, symbolOutput)
	}
	lowerSymbols := strings.ToLower(string(symbolOutput))
	for _, forbidden := range []string{"stm32", "initclk", "machine.tim"} {
		if strings.Contains(lowerSymbols, forbidden) {
			return Result{}, fmt.Errorf("inspect linked ELF symbols: board-specific symbol %q remains", forbidden)
		}
	}
	sourceDir := filepath.Join(workDir, "Source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create device package source: %w", err)
	}
	packagedELFSource := filepath.Join(sourceDir, "pdex.elf")
	if err := copyFile(elfPath, packagedELFSource); err != nil {
		return Result{}, fmt.Errorf("stage device ELF: %w", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "pdxinfo"), []byte(probePDXInfo), 0o644); err != nil {
		return Result{}, fmt.Errorf("write device pdxinfo: %w", err)
	}
	pdxPath := filepath.Join(workDir, "DeviceProbe.pdx")
	if err := runCommand(ctx, workDir, "package device probe", pdc,
		"-sdkpath", sdkPath, "-q", sourceDir, pdxPath); err != nil {
		return Result{}, err
	}
	packagedBinary := filepath.Join(pdxPath, "pdex.bin")
	if err := requireNonEmptyFile(packagedBinary); err != nil {
		return Result{}, fmt.Errorf("verify packaged device binary: %w", err)
	}
	if info, statErr := os.Stat(filepath.Join(pdxPath, "pdxinfo")); statErr != nil || info.IsDir() {
		return Result{}, fmt.Errorf("verify packaged pdxinfo: pdc did not create a file")
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
	deployment := "not requested"
	execution := "not requested"
	pending := "device deployment and hardware execution; allocator is non-collecting"
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
		if config.Run {
			runOutput, err := runDeviceProbe(ctx, workDir, pdutil)
			if err != nil {
				return Result{}, err
			}
			execution = summarizeRunOutput(runOutput)
			pending = "allocator is non-collecting"
		}
	}
	return Result{
		TinyGo:  firstLine(runVersion(ctx, tinygo, "version")),
		GCC:     firstLine(runVersion(ctx, gcc, "--version")),
		Format:  "ELF32 ARM executable, hard-float",
		Export:  "eventHandler",
		Package: "pdex.bin in .pdx",
		Deploy:  deployment,
		Run:     execution,
		Pending: pending,
	}, nil
}

func runDeviceProbe(ctx context.Context, directory, pdutil string) (string, error) {
	const attempts = 10
	for attempt := 1; attempt <= attempts; attempt++ {
		output, err := runCommandOutput(ctx, directory, "run device probe", pdutil, "run", deviceProbeInstallPath)
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

func runCommand(ctx context.Context, directory, action, executable string, args ...string) error {
	_, err := runCommandOutput(ctx, directory, action, executable, args...)
	return err
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

func renderProbeSource(modulePath string) string {
	return fmt.Sprintf(`package main

import (
	"unsafe"

	sdk %q
	sdkRuntime %q
	app %q
)

//go:linkname bridgeClear bridgeClear
func bridgeClear()

//go:linkname bridgeDrawText bridgeDrawText
func bridgeDrawText(text *byte, length uintptr, x, y int32)

type playdateContext struct{}

func (playdateContext) Clear() { bridgeClear() }

func (playdateContext) DrawText(text string, x, y int) {
	bridgeDrawText(unsafe.StringData(text), uintptr(len(text)), int32(x), int32(y))
}

var game sdk.Game = app.New()
var gameContext playdateContext
var gameRuntime = mustRuntime()

func mustRuntime() *sdkRuntime.Runtime {
	runtime, err := sdkRuntime.New(sdkRuntime.Callbacks{
		Init: func() error { return game.Init(gameContext) },
		Update: func() (bool, error) { return game.Update(gameContext) },
	})
	if err != nil {
		panic(err)
	}
	return runtime
}

//export goEventHandler
func goEventHandler(_ uintptr, event int32, arg uint32) int32 {
	if err := gameRuntime.Handle(sdkRuntime.Event(event), arg); err != nil {
		return -1
	}
	return 0
}

//export goUpdate
func goUpdate() int32 {
	refresh, err := gameRuntime.Update()
	if err != nil {
		return 0
	}
	return int32(refresh)
}

func main() {}
`, modulePath+"/playdate", modulePath+"/internal/features/runtime", modulePath+"/examples/hello")
}

const bootstrapSource = `#include "pd_api.h"

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
		playdate->system->setUpdateCallback(bridgeUpdate, playdate);
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

const linkerFlags = "-Wl,--gc-sections,--emit-relocs," +
	"--defsym,__exidx_start=0,--defsym,__exidx_end=0," +
	"--defsym,_sbss=__bss_start__,--defsym,_ebss=__bss_end__," +
	"--defsym,_sdata=__data_start__,--defsym,_edata=__data_end__,--defsym,_sidata=__etext," +
	"--defsym,_heap_start=__bss_end__,--defsym,_heap_end=__bss_end__," +
	"--defsym,_globals_start=__data_start__,--defsym,_globals_end=__bss_end__,--defsym,_stack_top=__bss_end__"

const probePDXInfo = `name=gopdsdk Device Probe
author=gopdsdk
bundleID=sdk.gopdsdk.deviceprobe
version=0.0.0
buildNumber=1
`

const deviceProbeInstallPath = "/Games/DeviceProbe.pdx"
