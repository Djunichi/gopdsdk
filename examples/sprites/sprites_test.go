package sprites

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Djunichi/gopdsdk/playdate"
)

type testBitmap struct{ operations *[]string }

func (*testBitmap) Width() (int, error)  { return 24, nil }
func (*testBitmap) Height() (int, error) { return 24, nil }
func (b *testBitmap) Clear() error       { return b.Fill(playdate.ColorClear) }
func (b *testBitmap) Fill(playdate.Color) error {
	*b.operations = append(*b.operations, "bitmap.fill")
	return nil
}
func (b *testBitmap) Close() error { *b.operations = append(*b.operations, "bitmap.close"); return nil }

type testSprite struct {
	name       string
	operations *[]string
}

func (s *testSprite) record(value string) error {
	*s.operations = append(*s.operations, s.name+"."+value)
	return nil
}
func (s *testSprite) SetBitmap(playdate.Bitmap) error    { return s.record("bitmap") }
func (s *testSprite) SetPosition(float32, float32) error { return s.record("position") }
func (s *testSprite) MoveBy(float32, float32) error      { return s.record("move") }
func (s *testSprite) SetVisible(bool) error              { return s.record("visible") }
func (s *testSprite) SetZIndex(int) error                { return s.record("z") }
func (s *testSprite) SetCollideRect(playdate.Rect) error { return s.record("collideRect") }
func (s *testSprite) ClearCollideRect() error            { return s.record("clearCollideRect") }
func (s *testSprite) SetTag(uint8) error                 { return s.record("tag") }
func (s *testSprite) MoveWithCollisions(float32, float32) (playdate.MoveResult, error) {
	return playdate.MoveResult{}, s.record("collideMove")
}
func (s *testSprite) Add() error    { return s.record("add") }
func (s *testSprite) Remove() error { return s.record("remove") }
func (s *testSprite) Close() error  { return s.record("close") }

type testContext struct {
	operations []string
	input      playdate.Input
	create     int
	failAt     int
}

func (*testContext) Clear()                                               {}
func (*testContext) DrawText(string, int, int)                            {}
func (*testContext) CurrentTimeMilliseconds() uint32                      { return 0 }
func (c *testContext) Input() playdate.Input                              { return c.input }
func (c *testContext) LoadBitmap(string) (playdate.Bitmap, error)         { return nil, nil }
func (*testContext) LoadBitmapTable(string) (playdate.BitmapTable, error) { return nil, nil }
func (c *testContext) NewBitmap(int, int) (playdate.Bitmap, error) {
	return &testBitmap{&c.operations}, nil
}
func (*testContext) DrawBitmap(playdate.Bitmap, int, int) error                         { return nil }
func (*testContext) DrawScaledBitmap(playdate.Bitmap, int, int, float32, float32) error { return nil }
func (c *testContext) NewSprite() (playdate.Sprite, error) {
	c.create++
	if c.create == c.failAt {
		return nil, errors.New("create failed")
	}
	return &testSprite{name: string(rune('0' + c.create)), operations: &c.operations}, nil
}
func (*testContext) QuerySpritesAtPoint(float32, float32) []playdate.Sprite { return nil }
func (*testContext) QuerySpritesInRect(playdate.Rect) []playdate.Sprite     { return nil }
func (*testContext) QueryOverlappingSprites(playdate.Sprite) ([]playdate.Sprite, error) {
	return nil, nil
}
func (c *testContext) UpdateAndDrawSprites() { c.operations = append(c.operations, "draw") }

func TestAcceptanceMovesCrankSpriteAndDrawsDisplayList(t *testing.T) {
	context := &testContext{input: playdate.Input{CrankDelta: 4}}
	game := New().(*game)
	if err := game.Init(context); err != nil {
		t.Fatal(err)
	}
	context.operations = nil
	if refresh, err := game.Update(context); err != nil || !refresh {
		t.Fatalf("Update() = %t, %v", refresh, err)
	}
	want := []string{"1.position", "2.move", "3.move", "draw"}
	if !reflect.DeepEqual(context.operations, want) {
		t.Fatalf("operations = %v, want %v", context.operations, want)
	}
}

func TestInitializationRollbackClosesSpritesBeforeBitmap(t *testing.T) {
	context := &testContext{failAt: 3}
	err := New().Init(context)
	if err == nil {
		t.Fatal("Init() succeeded")
	}
	wantTail := []string{"1.close", "2.close", "bitmap.close"}
	got := context.operations[len(context.operations)-len(wantTail):]
	if !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("rollback tail = %v, want %v", got, wantTail)
	}
}
