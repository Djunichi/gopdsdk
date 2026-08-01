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
uintptr_t bridgeLoadBitmapTable(const char* path, const char** error);
uintptr_t bridgeBitmapTableFrame(uintptr_t table, int index);
void bridgeFreeBitmapTable(uintptr_t table);
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
void bridgeSpriteSetCollideRect(uintptr_t sprite, float x, float y, float width, float height);
void bridgeSpriteClearCollideRect(uintptr_t sprite);
void bridgeSpriteSetTag(uintptr_t sprite, uint8_t tag);
void* bridgeSpriteMoveWithCollisions(uintptr_t sprite, float x, float y, float* actualX, float* actualY, int* count);
void bridgeCollisionInfo(void* collisions, int index, uintptr_t* other, int* response, int* overlaps, float* ti, float* values);
void bridgeFreeCollisionList(void* collisions);
void* bridgeQuerySpritesAtPoint(float x, float y, int* count);
void* bridgeQuerySpritesInRect(float x, float y, float width, float height, int* count);
void* bridgeOverlappingSprites(uintptr_t sprite, int* count);
uintptr_t bridgeSpriteListItem(void* sprites, int index);
void bridgeFreeSpriteList(void* sprites);
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
var bitmapTableDriver = sdkRuntime.BitmapTableDriver{
	Frame: func(table uintptr, index int) uintptr { return uintptr(C.bridgeBitmapTableFrame(C.uintptr_t(table), C.int(index))) },
	Free: func(table uintptr) { C.bridgeFreeBitmapTable(C.uintptr_t(table)) },
}

func (playdateContext) LoadBitmap(path string) (sdkPlaydate.Bitmap, error) {
	cPath := C.CString(path); defer C.free(unsafe.Pointer(cPath))
	var message *C.char
	handle := uintptr(C.bridgeLoadBitmap(cPath, (**C.char)(unsafe.Pointer(&message))))
	if handle == 0 { if message != nil { return nil, sdkPlaydate.BitmapLoadError(C.GoString(message)) }; return nil, sdkPlaydate.BitmapLoadError("unknown error") }
	return sdkRuntime.NewOwnedBitmap(handle, bitmapDriver), nil
}
func (playdateContext) LoadBitmapTable(path string) (sdkPlaydate.BitmapTable, error) {
	cPath := C.CString(path); defer C.free(unsafe.Pointer(cPath)); var message *C.char
	handle := uintptr(C.bridgeLoadBitmapTable(cPath, (**C.char)(unsafe.Pointer(&message))))
	if handle == 0 { if message != nil { return nil, sdkPlaydate.BitmapLoadError(C.GoString(message)) }; return nil, sdkPlaydate.BitmapLoadError("unknown error") }
	return sdkRuntime.NewOwnedBitmapTable(handle, bitmapTableDriver, bitmapDriver), nil
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
	SetCollideRect: func(sprite uintptr, rect sdkPlaydate.Rect) { C.bridgeSpriteSetCollideRect(C.uintptr_t(sprite), C.float(rect.X), C.float(rect.Y), C.float(rect.Width), C.float(rect.Height)) },
	ClearCollideRect: func(sprite uintptr) { C.bridgeSpriteClearCollideRect(C.uintptr_t(sprite)) },
	SetTag: func(sprite uintptr, tag uint8) { C.bridgeSpriteSetTag(C.uintptr_t(sprite), C.uint8_t(tag)) },
	MoveWithCollisions: moveWithCollisions,
	Add: func(sprite uintptr) { C.bridgeSpriteAdd(C.uintptr_t(sprite)) },
	Remove: func(sprite uintptr) { C.bridgeSpriteRemove(C.uintptr_t(sprite)) },
	Free: func(sprite uintptr) { C.bridgeFreeSprite(C.uintptr_t(sprite)) },
}
func moveWithCollisions(sprite uintptr, x, y float32) (float32, float32, []sdkRuntime.NativeCollision) {
	var actualX, actualY C.float; var count C.int
	list := C.bridgeSpriteMoveWithCollisions(C.uintptr_t(sprite), C.float(x), C.float(y), &actualX, &actualY, &count)
	defer C.bridgeFreeCollisionList(list)
	result := make([]sdkRuntime.NativeCollision, int(count))
	for index := range result { var other C.uintptr_t; var response, overlaps C.int; var ti C.float; var values [16]C.float; C.bridgeCollisionInfo(list, C.int(index), &other, &response, &overlaps, &ti, &values[0]); result[index] = sdkRuntime.NativeCollision{Other: uintptr(other), ResponseType: sdkPlaydate.CollisionResponse(response), Overlaps: overlaps != 0, Time: float32(ti), Move: sdkPlaydate.Point{X: float32(values[0]), Y: float32(values[1])}, Normal: sdkPlaydate.Point{X: float32(values[2]), Y: float32(values[3])}, Touch: sdkPlaydate.Point{X: float32(values[4]), Y: float32(values[5])}, SpriteRect: sdkPlaydate.Rect{X: float32(values[6]), Y: float32(values[7]), Width: float32(values[8]), Height: float32(values[9])}, OtherRect: sdkPlaydate.Rect{X: float32(values[10]), Y: float32(values[11]), Width: float32(values[12]), Height: float32(values[13])}} }
	return float32(actualX), float32(actualY), result
}
func (playdateContext) NewSprite() (sdkPlaydate.Sprite, error) {
	handle := uintptr(C.bridgeNewSprite()); if handle == 0 { return nil, sdkPlaydate.ErrSpriteCreate }
	return sdkRuntime.NewOwnedSprite(handle, spriteDriver), nil
}
func querySprites(list unsafe.Pointer, count C.int) []sdkPlaydate.Sprite { defer C.bridgeFreeSpriteList(list); handles := make([]uintptr, int(count)); for index := range handles { handles[index] = uintptr(C.bridgeSpriteListItem(list, C.int(index))) }; return sdkRuntime.BorrowedSprites(handles, spriteDriver) }
func (playdateContext) QuerySpritesAtPoint(x, y float32) []sdkPlaydate.Sprite { var count C.int; list := C.bridgeQuerySpritesAtPoint(C.float(x), C.float(y), &count); return querySprites(list, count) }
func (playdateContext) QuerySpritesInRect(rect sdkPlaydate.Rect) []sdkPlaydate.Sprite { var count C.int; list := C.bridgeQuerySpritesInRect(C.float(rect.X), C.float(rect.Y), C.float(rect.Width), C.float(rect.Height), &count); return querySprites(list, count) }
func (playdateContext) QueryOverlappingSprites(sprite sdkPlaydate.Sprite) ([]sdkPlaydate.Sprite, error) { handle, err := sdkRuntime.SpriteHandle(sprite); if err != nil { return nil, err }; var count C.int; list := C.bridgeOverlappingSprites(C.uintptr_t(handle), &count); return querySprites(list, count), nil }
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
uintptr_t bridgeLoadBitmapTable(const char* path, const char** error) { return (uintptr_t)bridgePlaydate->graphics->loadBitmapTable(path, error); }
uintptr_t bridgeBitmapTableFrame(uintptr_t table, int index) { return (uintptr_t)bridgePlaydate->graphics->getTableBitmap((LCDBitmapTable*)table, index); }
void bridgeFreeBitmapTable(uintptr_t table) { bridgePlaydate->graphics->freeBitmapTable((LCDBitmapTable*)table); }
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
void bridgeSpriteSetCollideRect(uintptr_t sprite, float x, float y, float width, float height) { bridgePlaydate->sprite->setCollideRect((LCDSprite*)sprite, (PDRect){x, y, width, height}); }
void bridgeSpriteClearCollideRect(uintptr_t sprite) { bridgePlaydate->sprite->clearCollideRect((LCDSprite*)sprite); }
void bridgeSpriteSetTag(uintptr_t sprite, uint8_t tag) { bridgePlaydate->sprite->setTag((LCDSprite*)sprite, tag); }
void* bridgeSpriteMoveWithCollisions(uintptr_t sprite, float x, float y, float* actualX, float* actualY, int* count) { return bridgePlaydate->sprite->moveWithCollisions((LCDSprite*)sprite, x, y, actualX, actualY, count); }
void bridgeCollisionInfo(void* collisions, int index, uintptr_t* other, int* response, int* overlaps, float* ti, float* v) { SpriteCollisionInfo* c = &((SpriteCollisionInfo*)collisions)[index]; *other=(uintptr_t)c->other; *response=c->responseType; *overlaps=c->overlaps; *ti=c->ti; v[0]=c->move.x; v[1]=c->move.y; v[2]=c->normal.x; v[3]=c->normal.y; v[4]=c->touch.x; v[5]=c->touch.y; v[6]=c->spriteRect.x; v[7]=c->spriteRect.y; v[8]=c->spriteRect.width; v[9]=c->spriteRect.height; v[10]=c->otherRect.x; v[11]=c->otherRect.y; v[12]=c->otherRect.width; v[13]=c->otherRect.height; }
void bridgeFreeCollisionList(void* collisions) { if (collisions) bridgePlaydate->system->realloc(collisions, 0); }
void* bridgeQuerySpritesAtPoint(float x, float y, int* count) { return bridgePlaydate->sprite->querySpritesAtPoint(x, y, count); }
void* bridgeQuerySpritesInRect(float x, float y, float width, float height, int* count) { return bridgePlaydate->sprite->querySpritesInRect(x, y, width, height, count); }
void* bridgeOverlappingSprites(uintptr_t sprite, int* count) { return bridgePlaydate->sprite->overlappingSprites((LCDSprite*)sprite, count); }
uintptr_t bridgeSpriteListItem(void* sprites, int index) { return (uintptr_t)((LCDSprite**)sprites)[index]; }
void bridgeFreeSpriteList(void* sprites) { if (sprites) bridgePlaydate->system->realloc(sprites, 0); }
void bridgeSpriteAdd(uintptr_t sprite) { bridgePlaydate->sprite->addSprite((LCDSprite*)sprite); }
void bridgeSpriteRemove(uintptr_t sprite) { bridgePlaydate->sprite->removeSprite((LCDSprite*)sprite); }
void bridgeUpdateAndDrawSprites(void) { bridgePlaydate->sprite->updateAndDrawSprites(); }
`, filepath.ToSlash(apiHeader))
}
