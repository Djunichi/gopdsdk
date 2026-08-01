// Package simabi renders the Playdate Simulator ABI bridge for Go runtimes.
package simabi

import (
	"fmt"
	"path/filepath"
)

const (
	// EventHandlerExport is the entry point required by the Playdate Simulator.
	EventHandlerExport = "eventHandler"
	// UpdateExport is the Go update callback called by the C bridge.
	UpdateExport = "goUpdate"
)

// Config identifies the paths and import used by generated bridge sources.
type Config struct {
	APIHeader         string
	RuntimeImport     string
	ApplicationImport string
}

// Sources contains the generated Go and C sides of the Simulator bridge.
type Sources struct {
	Go string
	C  string
}

// Render generates a c-shared Simulator bridge for the runtime package.
func Render(config Config) (Sources, error) {
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "APIHeader", value: config.APIHeader},
		{name: "RuntimeImport", value: config.RuntimeImport},
		{name: "ApplicationImport", value: config.ApplicationImport},
	} {
		if field.value == "" {
			return Sources{}, fmt.Errorf("render Simulator ABI: %s is required", field.name)
		}
	}

	return Sources{
		Go: renderGo(config),
		C:  renderC(config.APIHeader),
	}, nil
}

func renderGo(config Config) string {
	return fmt.Sprintf(`package main

/*
#include %q
#include <stdlib.h>
void bridgeRegisterUpdate(PlaydateAPI* playdate);
void bridgeClear(void);
void bridgeDrawText(const char* text, size_t length, int x, int y);
*/
import "C"
import (
	"unsafe"

	sdkRuntime %q
	app %q
)

var activePlaydate *C.PlaydateAPI
var gameContext playdateContext

var application = mustApplication()

func mustApplication() *sdkRuntime.Application {
	application, err := sdkRuntime.NewApplication(app.New(), gameContext, func() {
		C.bridgeRegisterUpdate(activePlaydate)
	})
	if err != nil {
		panic(err)
	}
	return application
}

//export eventHandler
func eventHandler(playdate *C.PlaydateAPI, event C.PDSystemEvent, arg C.uint32_t) C.int {
	if sdkRuntime.Event(event) == sdkRuntime.EventInit {
		activePlaydate = playdate
	} else if activePlaydate == nil {
		return -1
	}
	if err := application.Handle(sdkRuntime.Event(event), uint32(arg)); err != nil {
		return -1
	}
	return 0
}

//export goUpdate
func goUpdate() C.int {
	refresh, err := application.Update()
	if err != nil {
		return 0
	}
	return C.int(refresh)
}

type playdateContext struct{}

func (playdateContext) Clear() { C.bridgeClear() }

func (playdateContext) DrawText(text string, x, y int) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.bridgeDrawText(cText, C.size_t(len(text)), C.int(x), C.int(y))
}

func main() {}
`,
		filepath.ToSlash(config.APIHeader),
		config.RuntimeImport,
		config.ApplicationImport,
	)
}

func renderC(apiHeader string) string {
	return fmt.Sprintf(`#include %q
#include <stddef.h>
#include <string.h>

_Static_assert(sizeof(PDSystemEvent) <= 4, "PDSystemEvent must fit a 32-bit call slot");
_Static_assert(kEventMirrorEnded <= INT32_MAX, "PDSystemEvent values must fit int32_t");
_Static_assert(sizeof(uint32_t) == 4, "event argument must be 32-bit");
_Static_assert(sizeof(int) == 4, "Playdate callback result must be 32-bit");

extern int goUpdate(void);

static PlaydateAPI* bridgePlaydate;

static int bridgeUpdate(void* userdata)
{
	(void)userdata;
	return goUpdate();
}

void bridgeRegisterUpdate(PlaydateAPI* playdate)
{
	bridgePlaydate = playdate;
	playdate->system->setUpdateCallback(bridgeUpdate, NULL);
}

void bridgeDrawText(const char* text, size_t length, int x, int y)
{
	bridgePlaydate->graphics->drawText(text, length, kUTF8Encoding, x, y);
}

void bridgeClear(void)
{
	bridgePlaydate->graphics->clear(kColorWhite);
}
`, filepath.ToSlash(apiHeader))
}
