package playdate_test

import (
	"bytes"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestExportedAPIBaseline(t *testing.T) {
	t.Parallel()

	actual := exportedAPI(t)
	if actual != apiBaseline {
		t.Fatalf("exported playdate API changed (-baseline +current):\n%s", lineDiff(apiBaseline, actual))
	}
}

func exportedAPI(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate API snapshot test")
	}
	directory := filepath.Dir(filename)
	files, err := filepath.Glob(filepath.Join(directory, "*.go"))
	if err != nil {
		t.Fatal(err)
	}

	set := token.NewFileSet()
	parsed := make([]*ast.File, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, parseErr := parser.ParseFile(set, file, nil, 0)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", file, parseErr)
		}
		parsed = append(parsed, source)
	}

	config := types.Config{Importer: importer.Default()}
	pkg, err := config.Check("github.com/Djunichi/gopdsdk/playdate", set, parsed, nil)
	if err != nil {
		t.Fatalf("type-check playdate: %v", err)
	}
	qualifier := func(other *types.Package) string {
		if other == pkg {
			return ""
		}
		return other.Path()
	}

	var lines []string
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		if !object.Exported() {
			continue
		}
		lines = append(lines, types.ObjectString(object, qualifier))
		typeName, ok := object.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}
		for methodIndex := 0; methodIndex < named.NumMethods(); methodIndex++ {
			method := named.Method(methodIndex)
			if method.Exported() {
				lines = append(lines, types.ObjectString(method, qualifier))
			}
		}
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func lineDiff(baseline, current string) string {
	oldLines := strings.Split(strings.TrimSuffix(baseline, "\n"), "\n")
	newLines := strings.Split(strings.TrimSuffix(current, "\n"), "\n")
	oldSet := make(map[string]bool, len(oldLines))
	newSet := make(map[string]bool, len(newLines))
	for _, line := range oldLines {
		oldSet[line] = true
	}
	for _, line := range newLines {
		newSet[line] = true
	}

	var result bytes.Buffer
	for _, line := range oldLines {
		if !newSet[line] {
			result.WriteString("- ")
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	for _, line := range newLines {
		if !oldSet[line] {
			result.WriteString("+ ")
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}

const apiBaseline = `const ButtonA Buttons
const ButtonB Buttons
const ButtonDown Buttons
const ButtonLeft Buttons
const ButtonRight Buttons
const ButtonUp Buttons
const CollisionBounce CollisionResponse
const CollisionFreeze CollisionResponse
const CollisionOverlap CollisionResponse
const CollisionSlide CollisionResponse
const ColorBlack Color
const ColorClear Color
const ColorWhite Color
const DrawModeBlackTransparent DrawMode
const DrawModeCopy DrawMode
const DrawModeFillBlack DrawMode
const DrawModeFillWhite DrawMode
const DrawModeInverted DrawMode
const DrawModeNXOR DrawMode
const DrawModeWhiteTransparent DrawMode
const DrawModeXOR DrawMode
const LifecycleLock LifecycleEvent
const LifecycleLowPower LifecycleEvent
const LifecyclePause LifecycleEvent
const LifecycleResume LifecycleEvent
const LifecycleTerminate LifecycleEvent
const LifecycleUnlock LifecycleEvent
const PlaybackPaused PlaybackState
const PlaybackPlaying PlaybackState
const PlaybackStopped PlaybackState
func (*Animation).Bitmap() (Bitmap, error)
func (*Animation).Frame() int
func (*Animation).Pause()
func (*Animation).Paused() bool
func (*Animation).Resume()
func (*Animation).SetFixedFrame(frame int) error
func (*Animation).Update(deltaSeconds float32)
func (*Animation).UseDeltaTime()
func (AudioLoadError).Error() string
func (BitmapLoadError).Error() string
func (Buttons).Has(requested Buttons) bool
func (FontLoadError).Error() string
func (FontLoadError).Is(target error) bool
func (Paint).Components() (solid uint8, pattern [16]byte, patterned bool)
func NewAnimation(table BitmapTable, first int, count int, frameSeconds float32) (*Animation, error)
func PatternPaint(image [8]byte, mask [8]byte) Paint
func SolidPaint(color Color) (Paint, error)
func XORPaint() Paint
type Animation struct{table BitmapTable; first int; count int; frame int; frameSeconds float32; elapsed float32; paused bool; fixed bool}
type Audio interface{LoadFilePlayer(path string) (FilePlayer, error); LoadSoundEffect(path string) (SoundEffect, error)}
type AudioLoadError string
type Bitmap interface{Clear() error; Close() error; Fill(Color) error; Height() (int, error); Width() (int, error)}
type BitmapLoadError string
type BitmapTable interface{Close() error; Frame(index int) (Bitmap, error)}
type Buttons uint8
type Collision struct{Other Sprite; ResponseType CollisionResponse; Overlaps bool; Time float32; Move Point; Normal Point; Touch Point; SpriteRect Rect; OtherRect Rect}
type CollisionResponse uint8
type Color uint8
type Context interface{System; Graphics; InputReader; Sprites; Audio}
type DrawMode uint8
type FilePlayer interface{Close() error; Pause() error; Play() error; Resume() error; SetVolume(left float32, right float32) error; State() (PlaybackState, error); Stop() error; Volume() (left float32, right float32, err error)}
type Font interface{Close() error; Height() (int, error); TextWidth(text string) (int, error)}
type FontGraphics interface{DrawTextFont(font Font, text string, x int, y int) error; LoadFont(path string) (Font, error)}
type FontLoadError string
type Game interface{Init(Context) error; Update(Context) (refresh bool, err error)}
type Graphics interface{Clear(); DrawBitmap(bitmap Bitmap, x int, y int) error; DrawScaledBitmap(bitmap Bitmap, x int, y int, scaleX float32, scaleY float32) error; DrawText(text string, x int, y int); LoadBitmap(path string) (Bitmap, error); LoadBitmapTable(path string) (BitmapTable, error); NewBitmap(width int, height int) (Bitmap, error)}
type GraphicsState interface{ClearClipRect(); SetClipRect(x int, y int, width int, height int) error; SetDrawMode(mode DrawMode) error; SetDrawOffset(dx int, dy int)}
type Input struct{Buttons Buttons; Pressed Buttons; Released Buttons; Held Buttons; CrankAngle float32; CrankDelta float32; CrankDocked bool; CrankDockedThisFrame bool; CrankUndocked bool; DeltaSeconds float32}
type InputReader interface{Input() Input}
type LifecycleEvent uint8
type LifecycleGame interface{HandleLifecycle(Context, LifecycleEvent) error}
type MoveResult struct{ActualX float32; ActualY float32; Collisions []Collision}
type Paint struct{pattern [16]byte; solid Color; kind uint8}
type PlaybackState uint8
type Point struct{X float32; Y float32}
type PrimitiveGraphics interface{DrawEllipse(x int, y int, width int, height int, lineWidth int, startAngle float32, endAngle float32, paint Paint) error; DrawLine(x1 int, y1 int, x2 int, y2 int, width int, paint Paint) error; DrawRect(x int, y int, width int, height int, paint Paint) error; DrawTriangle(x1 int, y1 int, x2 int, y2 int, x3 int, y3 int, width int, paint Paint) error; FillEllipse(x int, y int, width int, height int, startAngle float32, endAngle float32, paint Paint) error; FillRect(x int, y int, width int, height int, paint Paint) error; FillTriangle(x1 int, y1 int, x2 int, y2 int, x3 int, y3 int, paint Paint) error}
type Rect struct{X float32; Y float32; Width float32; Height float32}
type SoundEffect interface{Close() error; Pause() error; Play() error; Resume() error; SetVolume(left float32, right float32) error; State() (PlaybackState, error); Stop() error; Volume() (left float32, right float32, err error)}
type Sprite interface{Add() error; ClearCollideRect() error; Close() error; MoveBy(dx float32, dy float32) error; MoveWithCollisions(goalX float32, goalY float32) (MoveResult, error); Remove() error; SetBitmap(Bitmap) error; SetCollideRect(Rect) error; SetPosition(x float32, y float32) error; SetTag(uint8) error; SetVisible(bool) error; SetZIndex(int) error}
type Sprites interface{NewSprite() (Sprite, error); QueryOverlappingSprites(Sprite) ([]Sprite, error); QuerySpritesAtPoint(x float32, y float32) []Sprite; QuerySpritesInRect(Rect) []Sprite; UpdateAndDrawSprites()}
type System interface{CurrentTimeMilliseconds() uint32}
var ErrAnimationConfig error
var ErrAudioClosed error
var ErrAudioCreate error
var ErrAudioPlay error
var ErrAudioVolume error
var ErrBitmapBorrowed error
var ErrBitmapClosed error
var ErrBitmapColor error
var ErrBitmapCreate error
var ErrBitmapFrameRange error
var ErrBitmapScale error
var ErrBitmapSize error
var ErrBitmapTableBorrowed error
var ErrBitmapTableClosed error
var ErrFontClosed error
var ErrFontInvalid error
var ErrFontLoad error
var ErrGraphicsColor error
var ErrGraphicsDrawMode error
var ErrGraphicsGeometry error
var ErrGraphicsUnavailable error
var ErrSpriteBorrowed error
var ErrSpriteClosed error
var ErrSpriteCreate error
`
