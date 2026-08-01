package simabi

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	header := filepath.Join(t.TempDir(), "SDK with spaces", "C_API", "pd_api.h")
	sources, err := Render(Config{
		APIHeader:         header,
		RuntimeImport:     "github.com/Djunichi/gopdsdk/internal/features/runtime",
		PlaydateImport:    "github.com/Djunichi/gopdsdk/playdate",
		ApplicationImport: "github.com/Djunichi/gopdsdk/probe/app",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"#include " + strconv.Quote(filepath.ToSlash(header)),
		`github.com/Djunichi/gopdsdk/probe/app`,
		"sdkRuntime.NewApplication(app.New(), gameContext",
		"application.Handle",
		"application.Update",
		"sdkPlaydate.Buttons(C.bridgeButtons())",
		"C.bridgeCrankAngle",
		"C.bridgeCrankDelta",
		"C.bridgeCrankDocked",
		"C.bridgeFrameDelta",
		"C.bridgeClear",
		"C.bridgeDrawText",
		"C.bridgeCurrentTimeMilliseconds",
		"C.bridgeLoadBitmap",
		"C.bridgeNewBitmap",
		"C.bridgeFreeBitmap",
		"C.bridgeBitmapSize",
		"C.bridgeFillBitmap",
		"C.bridgeDrawBitmap",
		"C.bridgeDrawScaledBitmap",
		"C.bridgeNewSprite",
		"C.bridgeSpriteSetBitmap",
		"C.bridgeSpriteMoveTo",
		"C.bridgeSpriteMoveBy",
		"C.bridgeSpriteSetVisible",
		"C.bridgeSpriteSetZIndex",
		"C.bridgeSpriteAdd",
		"C.bridgeSpriteRemove",
		"C.bridgeFreeSprite",
		"C.bridgeUpdateAndDrawSprites",
	} {
		if !strings.Contains(sources.Go, want) {
			t.Errorf("Go source does not contain %q:\n%s", want, sources.Go)
		}
	}
	for _, want := range []string{
		"#include " + strconv.Quote(filepath.ToSlash(header)),
		"setUpdateCallback(bridgeUpdate, NULL)",
		"graphics->clear(kColorWhite)",
		"drawText(text, length, kUTF8Encoding, x, y)",
		"system->getCurrentTimeMilliseconds()",
		"system->getButtonState",
		"system->getCrankAngle",
		"system->getCrankChange",
		"system->isCrankDocked",
		"system->getElapsedTime",
		"system->resetElapsedTime",
		"graphics->loadBitmap",
		"graphics->newBitmap",
		"graphics->freeBitmap",
		"graphics->getBitmapData",
		"graphics->clearBitmap",
		"graphics->drawBitmap",
		"graphics->drawScaledBitmap",
		"sprite->newSprite",
		"sprite->setImage",
		"sprite->moveTo",
		"sprite->moveBy",
		"sprite->setVisible",
		"sprite->setZIndex",
		"sprite->addSprite",
		"sprite->removeSprite",
		"sprite->freeSprite",
		"sprite->updateAndDrawSprites",
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
