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
	PlaydateImport    string
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
		{name: "PlaydateImport", value: config.PlaydateImport},
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
uint32_t bridgeCurrentTimeMilliseconds(void);
uint32_t bridgeButtons(void);
float bridgeCrankAngle(void);
float bridgeCrankDelta(void);
int bridgeCrankDocked(void);
float bridgeFrameDelta(void);
uintptr_t bridgeLoadBitmap(const char* path, const char** error);
uintptr_t bridgeNewBitmap(int width, int height);
void bridgeFreeBitmap(uintptr_t bitmap);
void bridgeBitmapSize(uintptr_t bitmap, int* width, int* height);
void bridgeFillBitmap(uintptr_t bitmap, int color);
void bridgeDrawBitmap(uintptr_t bitmap, int x, int y);
void bridgeDrawScaledBitmap(uintptr_t bitmap, int x, int y, float scaleX, float scaleY);
uintptr_t bridgeNewSprite(void);
void bridgeFreeSprite(uintptr_t sprite);
void bridgeSpriteSetBitmap(uintptr_t sprite, uintptr_t bitmap);
void bridgeSpriteMoveTo(uintptr_t sprite, float x, float y);
void bridgeSpriteMoveBy(uintptr_t sprite, float dx, float dy);
void bridgeSpriteSetVisible(uintptr_t sprite, int visible);
void bridgeSpriteSetZIndex(uintptr_t sprite, int z);
void bridgeSpriteAdd(uintptr_t sprite);
void bridgeSpriteRemove(uintptr_t sprite);
void bridgeUpdateAndDrawSprites(void);
*/
import "C"
import (
	"unsafe"

	sdkRuntime %q
	sdkPlaydate %q
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
	refresh, err := application.Update(sdkRuntime.RawInput{
		Buttons: sdkPlaydate.Buttons(C.bridgeButtons()), CrankAngle: float32(C.bridgeCrankAngle()),
		CrankDelta: float32(C.bridgeCrankDelta()), CrankDocked: C.bridgeCrankDocked() != 0,
		DeltaSeconds: float32(C.bridgeFrameDelta()),
	})
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

func (playdateContext) CurrentTimeMilliseconds() uint32 {
	return uint32(C.bridgeCurrentTimeMilliseconds())
}

func (playdateContext) Input() sdkPlaydate.Input { return sdkPlaydate.Input{} }

var bitmapDriver = sdkRuntime.BitmapDriver{
	Dimensions: func(handle uintptr) (int, int) { var width, height C.int; C.bridgeBitmapSize(C.uintptr_t(handle), &width, &height); return int(width), int(height) },
	Fill: func(handle uintptr, color sdkPlaydate.Color) { C.bridgeFillBitmap(C.uintptr_t(handle), C.int(color)) },
	Free: func(handle uintptr) { C.bridgeFreeBitmap(C.uintptr_t(handle)) },
}

func (playdateContext) LoadBitmap(path string) (sdkPlaydate.Bitmap, error) {
	cPath := C.CString(path); defer C.free(unsafe.Pointer(cPath))
	var message *C.char
	handle := uintptr(C.bridgeLoadBitmap(cPath, (**C.char)(unsafe.Pointer(&message))))
	if handle == 0 { if message != nil { return nil, sdkPlaydate.BitmapLoadError(C.GoString(message)) }; return nil, sdkPlaydate.BitmapLoadError("unknown error") }
	return sdkRuntime.NewOwnedBitmap(handle, bitmapDriver), nil
}
func (playdateContext) NewBitmap(width, height int) (sdkPlaydate.Bitmap, error) {
	if err := sdkRuntime.ValidateBitmapSize(width, height); err != nil { return nil, err }
	handle := uintptr(C.bridgeNewBitmap(C.int(width), C.int(height)))
	if handle == 0 { return nil, sdkPlaydate.ErrBitmapCreate }
	return sdkRuntime.NewOwnedBitmap(handle, bitmapDriver), nil
}
func (playdateContext) DrawBitmap(bitmap sdkPlaydate.Bitmap, x, y int) error {
	handle, err := sdkRuntime.BitmapHandle(bitmap); if err != nil { return err }
	C.bridgeDrawBitmap(C.uintptr_t(handle), C.int(x), C.int(y)); return nil
}
func (playdateContext) DrawScaledBitmap(bitmap sdkPlaydate.Bitmap, x, y int, scaleX, scaleY float32) error {
	if err := sdkRuntime.ValidateBitmapScale(scaleX, scaleY); err != nil { return err }
	handle, err := sdkRuntime.BitmapHandle(bitmap); if err != nil { return err }
	C.bridgeDrawScaledBitmap(C.uintptr_t(handle), C.int(x), C.int(y), C.float(scaleX), C.float(scaleY)); return nil
}

var spriteDriver = sdkRuntime.SpriteDriver{
	SetBitmap: func(sprite, bitmap uintptr) { C.bridgeSpriteSetBitmap(C.uintptr_t(sprite), C.uintptr_t(bitmap)) },
	MoveTo: func(sprite uintptr, x, y float32) { C.bridgeSpriteMoveTo(C.uintptr_t(sprite), C.float(x), C.float(y)) },
	MoveBy: func(sprite uintptr, dx, dy float32) { C.bridgeSpriteMoveBy(C.uintptr_t(sprite), C.float(dx), C.float(dy)) },
	SetVisible: func(sprite uintptr, visible bool) { value := C.int(0); if visible { value = 1 }; C.bridgeSpriteSetVisible(C.uintptr_t(sprite), value) },
	SetZIndex: func(sprite uintptr, z int) { C.bridgeSpriteSetZIndex(C.uintptr_t(sprite), C.int(z)) },
	Add: func(sprite uintptr) { C.bridgeSpriteAdd(C.uintptr_t(sprite)) },
	Remove: func(sprite uintptr) { C.bridgeSpriteRemove(C.uintptr_t(sprite)) },
	Free: func(sprite uintptr) { C.bridgeFreeSprite(C.uintptr_t(sprite)) },
}
func (playdateContext) NewSprite() (sdkPlaydate.Sprite, error) {
	handle := uintptr(C.bridgeNewSprite()); if handle == 0 { return nil, sdkPlaydate.ErrSpriteCreate }
	return sdkRuntime.NewOwnedSprite(handle, spriteDriver), nil
}
func (playdateContext) UpdateAndDrawSprites() { C.bridgeUpdateAndDrawSprites() }

func main() {}
`,
		filepath.ToSlash(config.APIHeader),
		config.RuntimeImport,
		config.PlaydateImport,
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
	playdate->system->resetElapsedTime();
}

void bridgeDrawText(const char* text, size_t length, int x, int y)
{
	bridgePlaydate->graphics->drawText(text, length, kUTF8Encoding, x, y);
}

void bridgeClear(void)
{
	bridgePlaydate->graphics->clear(kColorWhite);
}

uint32_t bridgeCurrentTimeMilliseconds(void)
{
	return bridgePlaydate->system->getCurrentTimeMilliseconds();
}

uint32_t bridgeButtons(void)
{
	PDButtons current = 0;
	bridgePlaydate->system->getButtonState(&current, NULL, NULL);
	return (uint32_t)current;
}

float bridgeCrankAngle(void) { return bridgePlaydate->system->getCrankAngle(); }
float bridgeCrankDelta(void) { return bridgePlaydate->system->getCrankChange(); }
int bridgeCrankDocked(void) { return bridgePlaydate->system->isCrankDocked(); }

float bridgeFrameDelta(void)
{
	float elapsed = bridgePlaydate->system->getElapsedTime();
	bridgePlaydate->system->resetElapsedTime();
	return elapsed;
}

uintptr_t bridgeLoadBitmap(const char* path, const char** error) { return (uintptr_t)bridgePlaydate->graphics->loadBitmap(path, error); }
uintptr_t bridgeNewBitmap(int width, int height) { return (uintptr_t)bridgePlaydate->graphics->newBitmap(width, height, kColorClear); }
void bridgeFreeBitmap(uintptr_t bitmap) { bridgePlaydate->graphics->freeBitmap((LCDBitmap*)bitmap); }
void bridgeBitmapSize(uintptr_t bitmap, int* width, int* height) { bridgePlaydate->graphics->getBitmapData((LCDBitmap*)bitmap, width, height, NULL, NULL, NULL); }
static LCDColor bridgeBitmapColor(int color) { return color == 1 ? kColorWhite : color == 2 ? kColorBlack : kColorClear; }
void bridgeFillBitmap(uintptr_t bitmap, int color) { bridgePlaydate->graphics->clearBitmap((LCDBitmap*)bitmap, bridgeBitmapColor(color)); }
void bridgeDrawBitmap(uintptr_t bitmap, int x, int y) { bridgePlaydate->graphics->drawBitmap((LCDBitmap*)bitmap, x, y, kBitmapUnflipped); }
void bridgeDrawScaledBitmap(uintptr_t bitmap, int x, int y, float scaleX, float scaleY) { bridgePlaydate->graphics->drawScaledBitmap((LCDBitmap*)bitmap, x, y, scaleX, scaleY); }
uintptr_t bridgeNewSprite(void) { return (uintptr_t)bridgePlaydate->sprite->newSprite(); }
void bridgeFreeSprite(uintptr_t sprite) { bridgePlaydate->sprite->freeSprite((LCDSprite*)sprite); }
void bridgeSpriteSetBitmap(uintptr_t sprite, uintptr_t bitmap) { bridgePlaydate->sprite->setImage((LCDSprite*)sprite, (LCDBitmap*)bitmap, kBitmapUnflipped); }
void bridgeSpriteMoveTo(uintptr_t sprite, float x, float y) { bridgePlaydate->sprite->moveTo((LCDSprite*)sprite, x, y); }
void bridgeSpriteMoveBy(uintptr_t sprite, float dx, float dy) { bridgePlaydate->sprite->moveBy((LCDSprite*)sprite, dx, dy); }
void bridgeSpriteSetVisible(uintptr_t sprite, int visible) { bridgePlaydate->sprite->setVisible((LCDSprite*)sprite, visible); }
void bridgeSpriteSetZIndex(uintptr_t sprite, int z) { bridgePlaydate->sprite->setZIndex((LCDSprite*)sprite, z); }
void bridgeSpriteAdd(uintptr_t sprite) { bridgePlaydate->sprite->addSprite((LCDSprite*)sprite); }
void bridgeSpriteRemove(uintptr_t sprite) { bridgePlaydate->sprite->removeSprite((LCDSprite*)sprite); }
void bridgeUpdateAndDrawSprites(void) { bridgePlaydate->sprite->updateAndDrawSprites(); }
`, filepath.ToSlash(apiHeader))
}
