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
const LifecycleLock LifecycleEvent
const LifecycleLowPower LifecycleEvent
const LifecyclePause LifecycleEvent
const LifecycleResume LifecycleEvent
const LifecycleTerminate LifecycleEvent
const LifecycleUnlock LifecycleEvent
func (BitmapLoadError).Error() string
func (Buttons).Has(requested Buttons) bool
type Bitmap interface{Clear() error; Close() error; Fill(Color) error; Height() (int, error); Width() (int, error)}
type BitmapLoadError string
type Buttons uint8
type Collision struct{Other Sprite; ResponseType CollisionResponse; Overlaps bool; Time float32; Move Point; Normal Point; Touch Point; SpriteRect Rect; OtherRect Rect}
type CollisionResponse uint8
type Color uint8
type Context interface{System; Graphics; InputReader; Sprites}
type Game interface{Init(Context) error; Update(Context) (refresh bool, err error)}
type Graphics interface{Clear(); DrawBitmap(bitmap Bitmap, x int, y int) error; DrawScaledBitmap(bitmap Bitmap, x int, y int, scaleX float32, scaleY float32) error; DrawText(text string, x int, y int); LoadBitmap(path string) (Bitmap, error); NewBitmap(width int, height int) (Bitmap, error)}
type Input struct{Buttons Buttons; Pressed Buttons; Released Buttons; Held Buttons; CrankAngle float32; CrankDelta float32; CrankDocked bool; CrankDockedThisFrame bool; CrankUndocked bool; DeltaSeconds float32}
type InputReader interface{Input() Input}
type LifecycleEvent uint8
type LifecycleGame interface{HandleLifecycle(Context, LifecycleEvent) error}
type MoveResult struct{ActualX float32; ActualY float32; Collisions []Collision}
type Point struct{X float32; Y float32}
type Rect struct{X float32; Y float32; Width float32; Height float32}
type Sprite interface{Add() error; ClearCollideRect() error; Close() error; MoveBy(dx float32, dy float32) error; MoveWithCollisions(goalX float32, goalY float32) (MoveResult, error); Remove() error; SetBitmap(Bitmap) error; SetCollideRect(Rect) error; SetPosition(x float32, y float32) error; SetTag(uint8) error; SetVisible(bool) error; SetZIndex(int) error}
type Sprites interface{NewSprite() (Sprite, error); QueryOverlappingSprites(Sprite) ([]Sprite, error); QuerySpritesAtPoint(x float32, y float32) []Sprite; QuerySpritesInRect(Rect) []Sprite; UpdateAndDrawSprites()}
type System interface{CurrentTimeMilliseconds() uint32}
var ErrBitmapBorrowed error
var ErrBitmapClosed error
var ErrBitmapColor error
var ErrBitmapCreate error
var ErrBitmapScale error
var ErrBitmapSize error
var ErrSpriteBorrowed error
var ErrSpriteClosed error
var ErrSpriteCreate error
`
