package deviceprobe

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Djunichi/gopdsdk/internal/shared/gomodule"
)

func TestRenderDeviceGoModAddsExternalApplicationModule(t *testing.T) {
	gameDir := filepath.Join(t.TempDir(), "game")
	app := applicationInfo{ImportPath: "example.com/game/pkg", Name: "pkg", Dir: filepath.Join(gameDir, "pkg")}
	app.Module = &struct{ Path, Dir, GoVersion string }{Path: "example.com/game", Dir: gameDir, GoVersion: "1.26"}
	got := renderDeviceGoMod(gomodule.Info{Path: "github.com/Djunichi/gopdsdk", Root: filepath.Join(t.TempDir(), "sdk"), GoVersion: "1.26"}, app)
	for _, want := range []string{"require example.com/game v0.0.0", "replace example.com/game =>", strconv.Quote(filepath.ToSlash(gameDir))} {
		if !strings.Contains(got, want) {
			t.Errorf("renderDeviceGoMod() does not contain %q:\n%s", want, got)
		}
	}
}

func TestProbeSourceExportsGoEventHandler(t *testing.T) {
	source := renderProbeSource("github.com/Djunichi/gopdsdk", "example.com/game")
	for _, want := range []string{"package main", "//export goEventHandler", "func goEventHandler", "//export goUpdate", "sdkRuntime.NewApplication(app.New(), gameContext, nil)", "application.Handle", "application.Update", "bridgeClear", "bridgeDrawText", `"example.com/game"`, "func main()"} {
		if !strings.Contains(source, want) {
			t.Errorf("probe source does not contain %q", want)
		}
	}
}

func TestBootstrapInitializesRuntimeOnce(t *testing.T) {
	for _, want := range []string{"runtime.run", "runtime.alloc", "activePlaydate->system->realloc(NULL, size)", "event == kEventInit && !booted", "runtimeRun();", "goEventHandler(playdate, event, arg)"} {
		if !strings.Contains(bootstrapSource, want) {
			t.Errorf("bootstrapSource does not contain %q", want)
		}
	}
	if strings.Contains(bootstrapSource, "runtime.preinit") || strings.Contains(bootstrapSource, "runtimePreinit") {
		t.Fatal("bootstrapSource must not call bare-metal preinit after the Playdate ELF loader")
	}
	activate := strings.Index(bootstrapSource, "activePlaydate = playdate;")
	run := strings.Index(bootstrapSource, "runtimeRun();")
	if activate < 0 || run < activate {
		t.Fatalf("bootstrap order activate/run = %d/%d", activate, run)
	}
}

func TestBootstrapDelegatesUpdateAndGraphicsToGo(t *testing.T) {
	for _, want := range []string{"result = goEventHandler(playdate, event, arg);", "setUpdateCallback(bridgeUpdate, playdate)", "return goUpdate();", "void bridgeClear(void)", "void bridgeDrawText", "graphics->drawText"} {
		if !strings.Contains(bootstrapSource, want) {
			t.Errorf("bootstrapSource does not contain %q", want)
		}
	}
}

func TestTargetUsesHardFloatCortexM7(t *testing.T) {
	for _, want := range []string{"cortex-m7", "thumbv7em-unknown-unknown-eabihf", `"relocation-model": "pic"`, `"qemu"`, "-mfloat-abi=hard", "-mfpu=fpv5-sp-d16"} {
		if !strings.Contains(targetSource, want) {
			t.Errorf("targetSource does not contain %q", want)
		}
	}
	for _, forbidden := range []string{"nucleof722ze", "stm32f722", "stm32f7", `"stm32"`} {
		if strings.Contains(targetSource, forbidden) {
			t.Errorf("targetSource contains board-specific tag %q", forbidden)
		}
	}
}

func TestAdapterDefinesInterruptHooks(t *testing.T) {
	for _, want := range []string{"DisableInterrupts:", "mrs r0, primask", "EnableInterrupts:", "msr primask, r0", "SemihostingCall:", "_exit:", "_kill:", "_getpid:", ".thumb_func"} {
		if !strings.Contains(adapterSource, want) {
			t.Errorf("adapterSource does not contain %q", want)
		}
	}
}

func TestFirstLine(t *testing.T) {
	if got, want := firstLine("first\r\nsecond\n"), "first"; got != want {
		t.Fatalf("firstLine() = %q, want %q", got, want)
	}
}

func TestRequireNonEmptyFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "pdex.bin")
	if err := os.WriteFile(path, []byte("device binary"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := requireNonEmptyFile(path); err != nil {
		t.Fatalf("requireNonEmptyFile() error = %v", err)
	}
}

func TestRequireNonEmptyFileRejectsEmptyFile(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "pdex.bin")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := requireNonEmptyFile(path); err == nil {
		t.Fatal("requireNonEmptyFile() error = nil, want empty-file error")
	}
}

func TestSummarizeOutput(t *testing.T) {
	if got, want := summarizeOutput("Playdate detected\r\nInstalled DeviceProbe.pdx\r\n"), "Playdate detected Installed DeviceProbe.pdx"; got != want {
		t.Fatalf("summarizeOutput() = %q, want %q", got, want)
	}
	if got, want := summarizeOutput("\r\n"), "installed by pdutil"; got != want {
		t.Fatalf("summarizeOutput(empty) = %q, want %q", got, want)
	}
}

func TestSummarizeRunOutput(t *testing.T) {
	if got, want := summarizeRunOutput("Playdate detected\r\nCommand sent.\r\n"), "Playdate detected Command sent."; got != want {
		t.Fatalf("summarizeRunOutput() = %q, want %q", got, want)
	}
	if got, want := summarizeRunOutput("\r\n"), "launch command sent by pdutil"; got != want {
		t.Fatalf("summarizeRunOutput(empty) = %q, want %q", got, want)
	}
}

func TestStrongUndefinedSymbolsIgnoresWeakReferences(t *testing.T) {
	output := "         w __gnu_Unwind_Find_exidx\n         U requiredSymbol\n"
	got := strongUndefinedSymbols(output)
	if len(got) != 1 || got[0] != "requiredSymbol" {
		t.Fatalf("strongUndefinedSymbols() = %v, want [requiredSymbol]", got)
	}
}

func TestUnsupportedRuntimeSymbolsRejectsDeferRuntime(t *testing.T) {
	output := "00003ae0 t runtime.setupDeferFrame\n000056b8 t runtime._recover\n00003298 t runtime/interrupt.In\n"
	got := unsupportedRuntimeSymbols(output)
	want := []string{"runtime.setupDeferFrame", "runtime._recover", "runtime/interrupt.In"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("unsupportedRuntimeSymbols() = %v, want %v", got, want)
	}
	if got := unsupportedRuntimeSymbols("00000100 t runtime.run\n"); len(got) != 0 {
		t.Fatalf("unsupportedRuntimeSymbols(safe) = %v, want none", got)
	}
}
