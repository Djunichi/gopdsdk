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
	APIHeader        string
	InitMarkerPath   string
	UpdateMarkerPath string
	RuntimeImport    string
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
		{name: "InitMarkerPath", value: config.InitMarkerPath},
		{name: "UpdateMarkerPath", value: config.UpdateMarkerPath},
		{name: "RuntimeImport", value: config.RuntimeImport},
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
void bridgeRegisterUpdate(PlaydateAPI* playdate);
void bridgeDrawHello(void);
*/
import "C"
import (
	"os"

	sdkRuntime %q
)

var activePlaydate *C.PlaydateAPI

var gameRuntime = mustRuntime()

func mustRuntime() *sdkRuntime.Runtime {
	runtime, err := sdkRuntime.New(sdkRuntime.Callbacks{
		Init: initGame,
		Update: updateGame,
	})
	if err != nil {
		panic(err)
	}
	return runtime
}

//export eventHandler
func eventHandler(playdate *C.PlaydateAPI, event C.PDSystemEvent, arg C.uint32_t) C.int {
	activePlaydate = playdate
	if err := gameRuntime.Handle(sdkRuntime.Event(event), uint32(arg)); err != nil {
		return -1
	}
	return 0
}

//export goUpdate
func goUpdate() C.int {
	refresh, err := gameRuntime.Update()
	if err != nil {
		return 0
	}
	return C.int(refresh)
}

func initGame() error {
	_ = os.WriteFile(%q, []byte("kEventInit"), 0o600)
	C.bridgeRegisterUpdate(activePlaydate)
	return nil
}

func updateGame() (bool, error) {
	C.bridgeDrawHello()
	_ = os.WriteFile(%q, []byte("update"), 0o600)
	return true, nil
}

func main() {}
`,
		filepath.ToSlash(config.APIHeader),
		config.RuntimeImport,
		filepath.ToSlash(config.InitMarkerPath),
		filepath.ToSlash(config.UpdateMarkerPath),
	)
}

func renderC(apiHeader string) string {
	return fmt.Sprintf(`#include %q
#include <stddef.h>
#include <string.h>

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

void bridgeDrawHello(void)
{
	static const char message[] = "Hello from gopdsdk";
	bridgePlaydate->graphics->drawText(message, strlen(message), kASCIIEncoding, 16, 16);
}
`, filepath.ToSlash(apiHeader))
}
