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
		"C.bridgeLoadFont",
		"C.bridgeTextWidth",
		"C.bridgeFontHeight",
		"C.bridgeFreeFont",
		"C.bridgeCurrentTimeMilliseconds",
		"C.bridgeExitToLauncher",
		"C.bridgeLoadBitmap",
		"C.bridgeLoadBitmapTable",
		"C.bridgeBitmapTableFrame",
		"C.bridgeFreeBitmapTable",
		"C.bridgeNewBitmap",
		"C.bridgeFreeBitmap",
		"C.bridgeBitmapSize",
		"C.bridgeFillBitmap",
		"C.bridgeDrawBitmap",
		"C.bridgeDrawScaledBitmap",
		"C.bridgeDrawLine",
		"C.bridgeDrawEllipse",
		"C.bridgeSetClipRect",
		"C.bridgeSetDrawMode",
		"C.bridgeGetFrame",
		"C.bridgeMarkUpdatedRows",
		"sdkRuntime.WithFramebuffer",
		"sdkRuntime.OwnedBitmapHandle",
		"C.bridgePushContext",
		"C.bridgePopContext",
		"C.bridgeNewSprite",
		"C.bridgeSpriteSetBitmap",
		"C.bridgeSpriteMoveTo",
		"C.bridgeSpriteMoveBy",
		"C.bridgeSpriteSetVisible",
		"C.bridgeSpriteSetZIndex",
		"C.bridgeSpriteSetCollideRect",
		"C.bridgeSpriteMoveWithCollisions",
		"C.bridgeQuerySpritesAtPoint",
		"C.bridgeQuerySpritesInRect",
		"C.bridgeSpriteAdd",
		"C.bridgeSpriteRemove",
		"C.bridgeFreeSprite",
		"C.bridgeUpdateAndDrawSprites",
		"C.bridgeLoadSoundEffect",
		"C.bridgeLoadFilePlayer",
		"sdkRuntime.NewSoundEffect",
		"sdkRuntime.NewFilePlayer",
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
		"graphics->loadFont",
		"graphics->getTextWidth",
		"graphics->getFontHeight",
		"system->realloc((void*)font, 0)",
		"system->getCurrentTimeMilliseconds()",
		"system->exitToLauncher()",
		"system->getButtonState",
		"system->getCrankAngle",
		"system->getCrankChange",
		"system->isCrankDocked",
		"system->getElapsedTime",
		"system->resetElapsedTime",
		"graphics->loadBitmap",
		"graphics->loadBitmapTable",
		"graphics->getTableBitmap",
		"graphics->freeBitmapTable",
		"graphics->newBitmap",
		"graphics->freeBitmap",
		"graphics->getBitmapData",
		"graphics->clearBitmap",
		"graphics->drawBitmap",
		"graphics->drawScaledBitmap",
		"graphics->drawLine",
		"graphics->fillTriangle",
		"graphics->setClipRect",
		"graphics->setDrawMode",
		"graphics->getFrame",
		"graphics->markUpdatedRows",
		"graphics->pushContext",
		"graphics->popContext",
		"sprite->newSprite",
		"sprite->setImage",
		"sprite->moveTo",
		"sprite->moveBy",
		"sprite->setVisible",
		"sprite->setZIndex",
		"sprite->setCollideRect",
		"sprite->moveWithCollisions",
		"sprite->querySpritesAtPoint",
		"sprite->querySpritesInRect",
		"sprite->addSprite",
		"sprite->removeSprite",
		"sprite->freeSprite",
		"sprite->updateAndDrawSprites",
		"sound->sample->load",
		"sound->sampleplayer->newPlayer",
		"sound->fileplayer->loadIntoPlayer",
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
