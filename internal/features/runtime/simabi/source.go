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
uint8_t* bridgeGetFrame(void);
void bridgeMarkUpdatedRows(int start, int end);
void bridgeDrawText(const char* text, size_t length, int x, int y);
uintptr_t bridgeLoadFont(const char* path, const char** error);
void bridgeSetFont(uintptr_t font);
int bridgeTextWidth(uintptr_t font, const char* text, size_t length);
int bridgeFontHeight(uintptr_t font);
void bridgeFreeFont(uintptr_t font);
uint32_t bridgeCurrentTimeMilliseconds(void);
void bridgeExitToLauncher(void);
const char* bridgeFileError(void);
uintptr_t bridgeFileOpen(const char* path, int options);
int bridgeFileClose(uintptr_t file);
int bridgeFileRead(uintptr_t file, void* buffer, unsigned int length);
int bridgeFileWrite(uintptr_t file, const void* buffer, unsigned int length);
int bridgeFileFlush(uintptr_t file);
int bridgeFileTell(uintptr_t file);
int bridgeFileSeek(uintptr_t file, int position, int whence);
int bridgeFileStat(const char* path, int* values);
uintptr_t bridgeFileList(const char* path, int showHidden, int* count, int* result);
const char* bridgeFileListItem(uintptr_t list, int index);
void bridgeFileListFree(uintptr_t list);
int bridgeFileMkdir(const char* path);
int bridgeFileRemove(const char* path, int recursive);
int bridgeFileRename(const char* from, const char* to);
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
void bridgeDrawLine(int x1, int y1, int x2, int y2, int width, int solid, const uint8_t* pattern, int patterned);
void bridgeDrawRect(int x, int y, int width, int height, int solid, const uint8_t* pattern, int patterned);
void bridgeFillRect(int x, int y, int width, int height, int solid, const uint8_t* pattern, int patterned);
void bridgeDrawEllipse(int x, int y, int width, int height, int lineWidth, float startAngle, float endAngle, int solid, const uint8_t* pattern, int patterned);
void bridgeFillEllipse(int x, int y, int width, int height, float startAngle, float endAngle, int solid, const uint8_t* pattern, int patterned);
void bridgeFillTriangle(int x1, int y1, int x2, int y2, int x3, int y3, int solid, const uint8_t* pattern, int patterned);
void bridgeDrawTriangle(int x1, int y1, int x2, int y2, int x3, int y3, int width, int solid, const uint8_t* pattern, int patterned);
void bridgeSetClipRect(int x, int y, int width, int height);
void bridgeClearClipRect(void);
void bridgeSetDrawOffset(int dx, int dy);
void bridgeSetDrawMode(int mode);
void bridgePushContext(uintptr_t bitmap);
void bridgePopContext(void);
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
uintptr_t bridgeLoadSoundEffect(const char* path);
int bridgeSoundEffectPlay(uintptr_t effect);
void bridgeSoundEffectStop(uintptr_t effect);
void bridgeSoundEffectSetVolume(uintptr_t effect, float left, float right);
void bridgeSoundEffectVolume(uintptr_t effect, float* left, float* right);
int bridgeSoundEffectIsPlaying(uintptr_t effect);
void bridgeSoundEffectPause(uintptr_t effect, int paused);
void bridgeFreeSoundEffect(uintptr_t effect);
uintptr_t bridgeLoadFilePlayer(const char* path);
int bridgeFilePlayerPlay(uintptr_t player);
void bridgeFilePlayerStop(uintptr_t player);
void bridgeFilePlayerSetVolume(uintptr_t player, float left, float right);
void bridgeFilePlayerVolume(uintptr_t player, float* left, float* right);
int bridgeFilePlayerIsPlaying(uintptr_t player);
void bridgeFilePlayerPause(uintptr_t player, int paused);
void bridgeFreeFilePlayer(uintptr_t player);
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

var _ sdkPlaydate.PrimitiveGraphics = playdateContext{}
var _ sdkPlaydate.GraphicsState = playdateContext{}
var _ sdkPlaydate.FramebufferGraphics = playdateContext{}
var _ sdkPlaydate.OffscreenGraphics = playdateContext{}
var _ sdkPlaydate.Launcher = playdateContext{}
var _ sdkPlaydate.FileSystem = playdateContext{}

func (playdateContext) Clear() { C.bridgeClear() }
func (playdateContext) WithFramebuffer(callback func(sdkPlaydate.Framebuffer) error) error {
	data := unsafe.Slice((*byte)(unsafe.Pointer(C.bridgeGetFrame())), 52*240)
	return sdkRuntime.WithFramebuffer(data, 400, 240, 52, func(start, end int) { C.bridgeMarkUpdatedRows(C.int(start), C.int(end)) }, callback)
}
func (playdateContext) DrawInto(bitmap sdkPlaydate.Bitmap, callback func() error) error {
	if callback == nil { return sdkPlaydate.ErrOffscreenCallback }
	handle, err := sdkRuntime.OwnedBitmapHandle(bitmap); if err != nil { return err }
	C.bridgePushContext(C.uintptr_t(handle)); err = callback(); C.bridgePopContext()
	return err
}

func (playdateContext) DrawText(text string, x, y int) {
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))
	C.bridgeDrawText(cText, C.size_t(len(text)), C.int(x), C.int(y))
}

var fontDriver = sdkRuntime.FontDriver{
	TextWidth: func(handle uintptr, text string) int { cText := C.CString(text); defer C.free(unsafe.Pointer(cText)); return int(C.bridgeTextWidth(C.uintptr_t(handle), cText, C.size_t(len(text)))) },
	Height: func(handle uintptr) int { return int(C.bridgeFontHeight(C.uintptr_t(handle))) },
	Free: func(handle uintptr) { C.bridgeFreeFont(C.uintptr_t(handle)) },
}

func (playdateContext) LoadFont(path string) (sdkPlaydate.Font, error) {
	cPath := C.CString(path); defer C.free(unsafe.Pointer(cPath)); var message *C.char
	handle := uintptr(C.bridgeLoadFont(cPath, (**C.char)(unsafe.Pointer(&message))))
	if handle == 0 { if message != nil { return nil, sdkPlaydate.FontLoadError(C.GoString(message)) }; return nil, sdkPlaydate.FontLoadError("unknown error") }
	return sdkRuntime.NewOwnedFont(handle, fontDriver), nil
}
func (playdateContext) DrawTextFont(font sdkPlaydate.Font, text string, x, y int) error {
	handle, err := sdkRuntime.FontHandle(font); if err != nil { return err }
	C.bridgeSetFont(C.uintptr_t(handle)); cText := C.CString(text); defer C.free(unsafe.Pointer(cText))
	C.bridgeDrawText(cText, C.size_t(len(text)), C.int(x), C.int(y)); return nil
}

func (playdateContext) CurrentTimeMilliseconds() uint32 {
	return uint32(C.bridgeCurrentTimeMilliseconds())
}

func (playdateContext) ExitToLauncher() { C.bridgeExitToLauncher() }

func fileErrorMessage() string { message := C.bridgeFileError(); if message == nil { return "" }; return C.GoString(message) }
var fileDriver = sdkRuntime.FileDriver{
	Read: func(handle uintptr, buffer []byte) (int, string) { result:=int(C.bridgeFileRead(C.uintptr_t(handle),unsafe.Pointer(unsafe.SliceData(buffer)),C.uint(len(buffer))));if result<0{return result,fileErrorMessage()};return result,"" },
	Write: func(handle uintptr, buffer []byte) (int, string) { result:=int(C.bridgeFileWrite(C.uintptr_t(handle),unsafe.Pointer(unsafe.SliceData(buffer)),C.uint(len(buffer))));if result<0{return result,fileErrorMessage()};return result,"" },
	Flush: func(handle uintptr) (int, string) { result:=int(C.bridgeFileFlush(C.uintptr_t(handle)));if result<0{return result,fileErrorMessage()};return result,"" },
	Tell: func(handle uintptr) (int, string) { result:=int(C.bridgeFileTell(C.uintptr_t(handle)));if result<0{return result,fileErrorMessage()};return result,"" },
	Seek: func(handle uintptr,position int32,whence int)(int,string){result:=int(C.bridgeFileSeek(C.uintptr_t(handle),C.int(position),C.int(whence)));if result<0{return result,fileErrorMessage()};return result,""},
	Close: func(handle uintptr)(int,string){result:=int(C.bridgeFileClose(C.uintptr_t(handle)));if result<0{return result,fileErrorMessage()};return result,""},
}
func (playdateContext) OpenFile(path string, options sdkPlaydate.FileOptions)(sdkPlaydate.File,error){if err:=sdkRuntime.ValidateFilePath(path,false);err!=nil{return nil,err};if err:=sdkRuntime.ValidateFileOptions(options);err!=nil{return nil,err};cPath:=C.CString(path);defer C.free(unsafe.Pointer(cPath));handle:=uintptr(C.bridgeFileOpen(cPath,C.int(options)));if handle==0{return nil,sdkPlaydate.FileOperationError{Operation:"open",Path:path,Message:fileErrorMessage()}};return sdkRuntime.NewOwnedFile(handle,path,fileDriver),nil}
func (playdateContext) Stat(path string)(sdkPlaydate.FileInfo,error){if err:=sdkRuntime.ValidateFilePath(path,false);err!=nil{return sdkPlaydate.FileInfo{},err};cPath:=C.CString(path);defer C.free(unsafe.Pointer(cPath));var values [8]C.int;if C.bridgeFileStat(cPath,&values[0])<0{return sdkPlaydate.FileInfo{},sdkPlaydate.FileOperationError{Operation:"stat",Path:path,Message:fileErrorMessage()}};return sdkPlaydate.FileInfo{IsDir:values[0]!=0,Size:uint32(values[1]),Year:int(values[2]),Month:int(values[3]),Day:int(values[4]),Hour:int(values[5]),Minute:int(values[6]),Second:int(values[7])},nil}
func (playdateContext) List(path string,showHidden bool)([]string,error){if err:=sdkRuntime.ValidateFilePath(path,true);err!=nil{return nil,err};cPath:=C.CString(path);defer C.free(unsafe.Pointer(cPath));hidden:=0;if showHidden{hidden=1};var count,result C.int;list:=uintptr(C.bridgeFileList(cPath,C.int(hidden),&count,&result));if result<0{message:="";if result == -1 {message=fileErrorMessage()};C.bridgeFileListFree(C.uintptr_t(list));return nil,sdkPlaydate.FileOperationError{Operation:"list",Path:path,Message:message}};defer C.bridgeFileListFree(C.uintptr_t(list));items:=make([]string,int(count));for i:=range items{items[i]=C.GoString(C.bridgeFileListItem(C.uintptr_t(list),C.int(i)))};return items,nil}
func filePathOperation(operation,path string,call func(*C.char)C.int)error{if err:=sdkRuntime.ValidateFilePath(path,false);err!=nil{return err};cPath:=C.CString(path);defer C.free(unsafe.Pointer(cPath));if call(cPath)<0{return sdkPlaydate.FileOperationError{Operation:operation,Path:path,Message:fileErrorMessage()}};return nil}
func (playdateContext) Mkdir(path string)error{return filePathOperation("mkdir",path,func(value *C.char)C.int{return C.bridgeFileMkdir(value)})}
func (playdateContext) Remove(path string,recursive bool)error{flag:=0;if recursive{flag=1};return filePathOperation("remove",path,func(value *C.char)C.int{return C.bridgeFileRemove(value,C.int(flag))})}
func (playdateContext) Rename(from,to string)error{if err:=sdkRuntime.ValidateFilePath(from,false);err!=nil{return err};if err:=sdkRuntime.ValidateFilePath(to,false);err!=nil{return err};cFrom:=C.CString(from);defer C.free(unsafe.Pointer(cFrom));cTo:=C.CString(to);defer C.free(unsafe.Pointer(cTo));if C.bridgeFileRename(cFrom,cTo)<0{return sdkPlaydate.FileOperationError{Operation:"rename",Path:from,Message:fileErrorMessage()}};return nil}

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

func paintArgs(paint sdkPlaydate.Paint) (C.int, *C.uint8_t, C.int, [16]byte) {
	solid, pattern, patterned := paint.Components(); flag := C.int(0); if patterned { flag = 1 }
	return C.int(solid), (*C.uint8_t)(unsafe.Pointer(&pattern[0])), flag, pattern
}
func (playdateContext) DrawLine(x1, y1, x2, y2, width int, paint sdkPlaydate.Paint) error {
	if err := sdkRuntime.ValidatePrimitiveGeometry(width, 1, 1, 0, 0); err != nil { return err }; solid, pattern, flag, keep := paintArgs(paint); _ = keep
	C.bridgeDrawLine(C.int(x1), C.int(y1), C.int(x2), C.int(y2), C.int(width), solid, pattern, flag); return nil
}
func (playdateContext) DrawRect(x, y, width, height int, paint sdkPlaydate.Paint) error { if err := sdkRuntime.ValidatePrimitiveGeometry(width, height, 1, 0, 0); err != nil { return err }; solid, pattern, flag, keep := paintArgs(paint); _ = keep; C.bridgeDrawRect(C.int(x), C.int(y), C.int(width), C.int(height), solid, pattern, flag); return nil }
func (playdateContext) FillRect(x, y, width, height int, paint sdkPlaydate.Paint) error { if err := sdkRuntime.ValidatePrimitiveGeometry(width, height, 1, 0, 0); err != nil { return err }; solid, pattern, flag, keep := paintArgs(paint); _ = keep; C.bridgeFillRect(C.int(x), C.int(y), C.int(width), C.int(height), solid, pattern, flag); return nil }
func (playdateContext) DrawEllipse(x, y, width, height, lineWidth int, startAngle, endAngle float32, paint sdkPlaydate.Paint) error { if err := sdkRuntime.ValidatePrimitiveGeometry(width, height, lineWidth, startAngle, endAngle); err != nil { return err }; solid, pattern, flag, keep := paintArgs(paint); _ = keep; C.bridgeDrawEllipse(C.int(x), C.int(y), C.int(width), C.int(height), C.int(lineWidth), C.float(startAngle), C.float(endAngle), solid, pattern, flag); return nil }
func (playdateContext) FillEllipse(x, y, width, height int, startAngle, endAngle float32, paint sdkPlaydate.Paint) error { if err := sdkRuntime.ValidatePrimitiveGeometry(width, height, 1, startAngle, endAngle); err != nil { return err }; solid, pattern, flag, keep := paintArgs(paint); _ = keep; C.bridgeFillEllipse(C.int(x), C.int(y), C.int(width), C.int(height), C.float(startAngle), C.float(endAngle), solid, pattern, flag); return nil }
func (playdateContext) FillTriangle(x1, y1, x2, y2, x3, y3 int, paint sdkPlaydate.Paint) error { solid, pattern, flag, keep := paintArgs(paint); _ = keep; C.bridgeFillTriangle(C.int(x1), C.int(y1), C.int(x2), C.int(y2), C.int(x3), C.int(y3), solid, pattern, flag); return nil }
func (playdateContext) DrawTriangle(x1, y1, x2, y2, x3, y3, width int, paint sdkPlaydate.Paint) error { if err := sdkRuntime.ValidatePrimitiveGeometry(width, 1, 1, 0, 0); err != nil { return err }; solid, pattern, flag, keep := paintArgs(paint); _ = keep; C.bridgeDrawTriangle(C.int(x1), C.int(y1), C.int(x2), C.int(y2), C.int(x3), C.int(y3), C.int(width), solid, pattern, flag); return nil }
func (playdateContext) SetClipRect(x, y, width, height int) error { if err := sdkRuntime.ValidatePrimitiveGeometry(width, height, 1, 0, 0); err != nil { return err }; C.bridgeSetClipRect(C.int(x), C.int(y), C.int(width), C.int(height)); return nil }
func (playdateContext) ClearClipRect() { C.bridgeClearClipRect() }
func (playdateContext) SetDrawOffset(dx, dy int) { C.bridgeSetDrawOffset(C.int(dx), C.int(dy)) }
func (playdateContext) SetDrawMode(mode sdkPlaydate.DrawMode) error { if err := sdkRuntime.ValidateDrawMode(mode); err != nil { return err }; C.bridgeSetDrawMode(C.int(mode)); return nil }

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

var soundEffectDriver = sdkRuntime.AudioDriver{
	Play: func(handle uintptr) bool { return C.bridgeSoundEffectPlay(C.uintptr_t(handle)) != 0 },
	Stop: func(handle uintptr) { C.bridgeSoundEffectStop(C.uintptr_t(handle)) },
	SetVolume: func(handle uintptr, left, right float32) { C.bridgeSoundEffectSetVolume(C.uintptr_t(handle), C.float(left), C.float(right)) },
	Volume: func(handle uintptr) (float32, float32) { var left, right C.float; C.bridgeSoundEffectVolume(C.uintptr_t(handle), &left, &right); return float32(left), float32(right) },
	IsPlaying: func(handle uintptr) bool { return C.bridgeSoundEffectIsPlaying(C.uintptr_t(handle)) != 0 },
	Pause: func(handle uintptr, paused bool) { value := C.int(0); if paused { value = 1 }; C.bridgeSoundEffectPause(C.uintptr_t(handle), value) },
	Free: func(handle uintptr) { C.bridgeFreeSoundEffect(C.uintptr_t(handle)) },
}
var filePlayerDriver = sdkRuntime.AudioDriver{
	Play: func(handle uintptr) bool { return C.bridgeFilePlayerPlay(C.uintptr_t(handle)) != 0 },
	Stop: func(handle uintptr) { C.bridgeFilePlayerStop(C.uintptr_t(handle)) },
	SetVolume: func(handle uintptr, left, right float32) { C.bridgeFilePlayerSetVolume(C.uintptr_t(handle), C.float(left), C.float(right)) },
	Volume: func(handle uintptr) (float32, float32) { var left, right C.float; C.bridgeFilePlayerVolume(C.uintptr_t(handle), &left, &right); return float32(left), float32(right) },
	IsPlaying: func(handle uintptr) bool { return C.bridgeFilePlayerIsPlaying(C.uintptr_t(handle)) != 0 },
	Pause: func(handle uintptr, _ bool) { C.bridgeFilePlayerPause(C.uintptr_t(handle), 1) },
	Free: func(handle uintptr) { C.bridgeFreeFilePlayer(C.uintptr_t(handle)) },
}
func (playdateContext) LoadSoundEffect(path string) (sdkPlaydate.SoundEffect, error) {
	cPath := C.CString(path); defer C.free(unsafe.Pointer(cPath)); handle := uintptr(C.bridgeLoadSoundEffect(cPath))
	if handle == 0 { return nil, sdkPlaydate.AudioLoadError(path) }
	return sdkRuntime.NewSoundEffect(handle, soundEffectDriver), nil
}
func (playdateContext) LoadFilePlayer(path string) (sdkPlaydate.FilePlayer, error) {
	cPath := C.CString(path); defer C.free(unsafe.Pointer(cPath)); handle := uintptr(C.bridgeLoadFilePlayer(cPath))
	if handle == 0 { return nil, sdkPlaydate.AudioLoadError(path) }
	return sdkRuntime.NewFilePlayer(handle, filePlayerDriver), nil
}

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

uintptr_t bridgeLoadFont(const char* path, const char** error) { return (uintptr_t)bridgePlaydate->graphics->loadFont(path, error); }
void bridgeSetFont(uintptr_t font) { bridgePlaydate->graphics->setFont((LCDFont*)font); }
int bridgeTextWidth(uintptr_t font, const char* text, size_t length) { return bridgePlaydate->graphics->getTextWidth((LCDFont*)font, text, length, kUTF8Encoding, 0); }
int bridgeFontHeight(uintptr_t font) { return bridgePlaydate->graphics->getFontHeight((LCDFont*)font); }
void bridgeFreeFont(uintptr_t font) { bridgePlaydate->system->realloc((void*)font, 0); }

void bridgeClear(void)
{
	bridgePlaydate->graphics->clear(kColorWhite);
}

uint8_t* bridgeGetFrame(void) { return bridgePlaydate->graphics->getFrame(); }
void bridgeMarkUpdatedRows(int start, int end) { bridgePlaydate->graphics->markUpdatedRows(start, end); }

uint32_t bridgeCurrentTimeMilliseconds(void)
{
	return bridgePlaydate->system->getCurrentTimeMilliseconds();
}

void bridgeExitToLauncher(void) { bridgePlaydate->system->exitToLauncher(); }
const char* bridgeFileError(void){return bridgePlaydate->file->geterr();}
uintptr_t bridgeFileOpen(const char* path,int options){return(uintptr_t)bridgePlaydate->file->open(path,(FileOptions)options);}
int bridgeFileClose(uintptr_t file){return bridgePlaydate->file->close((SDFile*)file);}
int bridgeFileRead(uintptr_t file,void* buffer,unsigned int length){return bridgePlaydate->file->read((SDFile*)file,buffer,length);}
int bridgeFileWrite(uintptr_t file,const void* buffer,unsigned int length){return bridgePlaydate->file->write((SDFile*)file,buffer,length);}
int bridgeFileFlush(uintptr_t file){return bridgePlaydate->file->flush((SDFile*)file);}
int bridgeFileTell(uintptr_t file){return bridgePlaydate->file->tell((SDFile*)file);}
int bridgeFileSeek(uintptr_t file,int position,int whence){return bridgePlaydate->file->seek((SDFile*)file,position,whence);}
int bridgeFileStat(const char* path,int* values){FileStat value;int result=bridgePlaydate->file->stat(path,&value);if(result<0)return result;values[0]=value.isdir;values[1]=(int)value.size;values[2]=value.m_year;values[3]=value.m_month;values[4]=value.m_day;values[5]=value.m_hour;values[6]=value.m_minute;values[7]=value.m_second;return 0;}
typedef struct{char** items;int count;int failed;}BridgeFileList;
static void bridgeCollectFile(const char* name,void* userdata){BridgeFileList* list=userdata;if(list->failed)return;char* copy=bridgePlaydate->system->realloc(NULL,strlen(name)+1);if(!copy){list->failed=1;return;}strcpy(copy,name);char** items=bridgePlaydate->system->realloc(list->items,sizeof(char*)*(list->count+1));if(!items){bridgePlaydate->system->realloc(copy,0);list->failed=1;return;}list->items=items;list->items[list->count++]=copy;}
uintptr_t bridgeFileList(const char* path,int showHidden,int* count,int* result){BridgeFileList* list=bridgePlaydate->system->realloc(NULL,sizeof(BridgeFileList));if(!list){*count=0;*result=-2;return 0;}list->items=NULL;list->count=0;list->failed=0;*result=bridgePlaydate->file->listfiles(path,bridgeCollectFile,list,showHidden);if(list->failed)*result=-2;*count=list->count;return(uintptr_t)list;}
const char* bridgeFileListItem(uintptr_t list,int index){return((BridgeFileList*)list)->items[index];}
void bridgeFileListFree(uintptr_t value){BridgeFileList* list=(BridgeFileList*)value;if(!list)return;for(int i=0;i<list->count;i++)bridgePlaydate->system->realloc(list->items[i],0);bridgePlaydate->system->realloc(list->items,0);bridgePlaydate->system->realloc(list,0);}
int bridgeFileMkdir(const char* path){return bridgePlaydate->file->mkdir(path);}
int bridgeFileRemove(const char* path,int recursive){return bridgePlaydate->file->unlink(path,recursive);}
int bridgeFileRename(const char* from,const char* to){return bridgePlaydate->file->rename(from,to);}

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
static LCDColor bridgePaint(int solid, const uint8_t* pattern, int patterned) { if (patterned) return (LCDColor)pattern; return solid == 1 ? kColorWhite : solid == 2 ? kColorBlack : solid == 3 ? kColorXOR : kColorClear; }
void bridgeDrawLine(int x1, int y1, int x2, int y2, int width, int solid, const uint8_t* pattern, int patterned) { bridgePlaydate->graphics->drawLine(x1, y1, x2, y2, width, bridgePaint(solid, pattern, patterned)); }
void bridgeDrawRect(int x, int y, int width, int height, int solid, const uint8_t* pattern, int patterned) { bridgePlaydate->graphics->drawRect(x, y, width, height, bridgePaint(solid, pattern, patterned)); }
void bridgeFillRect(int x, int y, int width, int height, int solid, const uint8_t* pattern, int patterned) { bridgePlaydate->graphics->fillRect(x, y, width, height, bridgePaint(solid, pattern, patterned)); }
void bridgeDrawEllipse(int x, int y, int width, int height, int lineWidth, float startAngle, float endAngle, int solid, const uint8_t* pattern, int patterned) { bridgePlaydate->graphics->drawEllipse(x, y, width, height, lineWidth, startAngle, endAngle, bridgePaint(solid, pattern, patterned)); }
void bridgeFillEllipse(int x, int y, int width, int height, float startAngle, float endAngle, int solid, const uint8_t* pattern, int patterned) { bridgePlaydate->graphics->fillEllipse(x, y, width, height, startAngle, endAngle, bridgePaint(solid, pattern, patterned)); }
void bridgeFillTriangle(int x1, int y1, int x2, int y2, int x3, int y3, int solid, const uint8_t* pattern, int patterned) { bridgePlaydate->graphics->fillTriangle(x1, y1, x2, y2, x3, y3, bridgePaint(solid, pattern, patterned)); }
void bridgeDrawTriangle(int x1, int y1, int x2, int y2, int x3, int y3, int width, int solid, const uint8_t* pattern, int patterned) { LCDColor color=bridgePaint(solid,pattern,patterned); bridgePlaydate->graphics->drawLine(x1,y1,x2,y2,width,color); bridgePlaydate->graphics->drawLine(x2,y2,x3,y3,width,color); bridgePlaydate->graphics->drawLine(x3,y3,x1,y1,width,color); }
void bridgeSetClipRect(int x, int y, int width, int height) { bridgePlaydate->graphics->setClipRect(x, y, width, height); }
void bridgeClearClipRect(void) { bridgePlaydate->graphics->clearClipRect(); }
void bridgeSetDrawOffset(int dx, int dy) { bridgePlaydate->graphics->setDrawOffset(dx, dy); }
void bridgeSetDrawMode(int mode) { bridgePlaydate->graphics->setDrawMode((LCDBitmapDrawMode)mode); }
void bridgePushContext(uintptr_t bitmap) { bridgePlaydate->graphics->pushContext((LCDBitmap*)bitmap); }
void bridgePopContext(void) { bridgePlaydate->graphics->popContext(); }
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

typedef struct { AudioSample* sample; SamplePlayer* player; } BridgeSoundEffect;
uintptr_t bridgeLoadSoundEffect(const char* path)
{
	AudioSample* sample = bridgePlaydate->sound->sample->load(path);
	if (!sample) return 0;
	SamplePlayer* player = bridgePlaydate->sound->sampleplayer->newPlayer();
	if (!player) { bridgePlaydate->sound->sample->freeSample(sample); return 0; }
	BridgeSoundEffect* effect = bridgePlaydate->system->realloc(NULL, sizeof(BridgeSoundEffect));
	if (!effect) { bridgePlaydate->sound->sampleplayer->freePlayer(player); bridgePlaydate->sound->sample->freeSample(sample); return 0; }
	effect->sample = sample; effect->player = player;
	bridgePlaydate->sound->sampleplayer->setSample(player, sample);
	return (uintptr_t)effect;
}
static BridgeSoundEffect* bridgeEffect(uintptr_t effect) { return (BridgeSoundEffect*)effect; }
int bridgeSoundEffectPlay(uintptr_t effect) { return bridgePlaydate->sound->sampleplayer->play(bridgeEffect(effect)->player, 1, 1.0f); }
void bridgeSoundEffectStop(uintptr_t effect) { bridgePlaydate->sound->sampleplayer->stop(bridgeEffect(effect)->player); }
void bridgeSoundEffectSetVolume(uintptr_t effect, float left, float right) { bridgePlaydate->sound->sampleplayer->setVolume(bridgeEffect(effect)->player, left, right); }
void bridgeSoundEffectVolume(uintptr_t effect, float* left, float* right) { bridgePlaydate->sound->sampleplayer->getVolume(bridgeEffect(effect)->player, left, right); }
int bridgeSoundEffectIsPlaying(uintptr_t effect) { return bridgePlaydate->sound->sampleplayer->isPlaying(bridgeEffect(effect)->player); }
void bridgeSoundEffectPause(uintptr_t effect, int paused) { bridgePlaydate->sound->sampleplayer->setPaused(bridgeEffect(effect)->player, paused); }
void bridgeFreeSoundEffect(uintptr_t effect) { BridgeSoundEffect* value = bridgeEffect(effect); bridgePlaydate->sound->sampleplayer->freePlayer(value->player); bridgePlaydate->sound->sample->freeSample(value->sample); bridgePlaydate->system->realloc(value, 0); }

uintptr_t bridgeLoadFilePlayer(const char* path) { FilePlayer* player = bridgePlaydate->sound->fileplayer->newPlayer(); if (!player) return 0; if (!bridgePlaydate->sound->fileplayer->loadIntoPlayer(player, path)) { bridgePlaydate->sound->fileplayer->freePlayer(player); return 0; } return (uintptr_t)player; }
int bridgeFilePlayerPlay(uintptr_t player) { return bridgePlaydate->sound->fileplayer->play((FilePlayer*)player, 1); }
void bridgeFilePlayerStop(uintptr_t player) { bridgePlaydate->sound->fileplayer->stop((FilePlayer*)player); }
void bridgeFilePlayerSetVolume(uintptr_t player, float left, float right) { bridgePlaydate->sound->fileplayer->setVolume((FilePlayer*)player, left, right); }
void bridgeFilePlayerVolume(uintptr_t player, float* left, float* right) { bridgePlaydate->sound->fileplayer->getVolume((FilePlayer*)player, left, right); }
int bridgeFilePlayerIsPlaying(uintptr_t player) { return bridgePlaydate->sound->fileplayer->isPlaying((FilePlayer*)player); }
void bridgeFilePlayerPause(uintptr_t player, int paused) { (void)paused; bridgePlaydate->sound->fileplayer->pause((FilePlayer*)player); }
void bridgeFreeFilePlayer(uintptr_t player) { bridgePlaydate->sound->fileplayer->freePlayer((FilePlayer*)player); }
`, filepath.ToSlash(apiHeader))
}
