package simabi

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	sources, err := Render(Config{
		APIHeader:         `C:\SDK with spaces\C_API\pd_api.h`,
		PublicAPIImport:   "github.com/Djunichi/gopdsdk/playdate",
		RuntimeImport:     "github.com/Djunichi/gopdsdk/internal/features/runtime",
		ApplicationImport: "github.com/Djunichi/gopdsdk/probe/app",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`#include "C:/SDK with spaces/C_API/pd_api.h"`,
		`github.com/Djunichi/gopdsdk/probe/app`,
		`github.com/Djunichi/gopdsdk/playdate`,
		"var game sdk.Game = app.New()",
		"gameRuntime.Handle",
		"gameRuntime.Update",
		"game.Init(gameContext)",
		"game.Update(gameContext)",
		"C.bridgeClear",
		"C.bridgeDrawText",
	} {
		if !strings.Contains(sources.Go, want) {
			t.Errorf("Go source does not contain %q:\n%s", want, sources.Go)
		}
	}
	for _, want := range []string{
		`#include "C:/SDK with spaces/C_API/pd_api.h"`,
		"setUpdateCallback(bridgeUpdate, NULL)",
		"graphics->clear(kColorWhite)",
		"drawText(text, length, kUTF8Encoding, x, y)",
	} {
		if !strings.Contains(sources.C, want) {
			t.Errorf("C source does not contain %q:\n%s", want, sources.C)
		}
	}
}

func TestRenderRequiresEveryInput(t *testing.T) {
	_, err := Render(Config{})
	if got, want := err.Error(), "render Simulator ABI: APIHeader is required"; got != want {
		t.Fatalf("Render() error = %q, want %q", got, want)
	}
}
