package simabi

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	sources, err := Render(Config{
		APIHeader:        `C:\SDK with spaces\C_API\pd_api.h`,
		InitMarkerPath:   `C:\Temp path\init.marker`,
		UpdateMarkerPath: `C:\Temp path\update.marker`,
		RuntimeImport:    "github.com/Djunichi/gopdsdk/internal/features/runtime",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`#include "C:/SDK with spaces/C_API/pd_api.h"`,
		`"C:/Temp path/init.marker"`,
		`"C:/Temp path/update.marker"`,
		"gameRuntime.Handle",
		"gameRuntime.Update",
	} {
		if !strings.Contains(sources.Go, want) {
			t.Errorf("Go source does not contain %q:\n%s", want, sources.Go)
		}
	}
	for _, want := range []string{
		`#include "C:/SDK with spaces/C_API/pd_api.h"`,
		"setUpdateCallback(bridgeUpdate, NULL)",
		"drawText(message, strlen(message), kASCIIEncoding, 16, 16)",
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
