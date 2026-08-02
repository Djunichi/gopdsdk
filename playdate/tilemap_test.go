package playdate

import (
	"errors"
	"reflect"
	"testing"
)

type tileBitmap struct{}

func (*tileBitmap) Width() (int, error)  { return 16, nil }
func (*tileBitmap) Height() (int, error) { return 16, nil }
func (*tileBitmap) Clear() error         { return nil }
func (*tileBitmap) Fill(Color) error     { return nil }
func (*tileBitmap) Close() error         { return nil }

type tileGraphics struct {
	positions [][2]int
	fail      error
}

func (*tileGraphics) Clear()                                      {}
func (*tileGraphics) DrawText(string, int, int)                   {}
func (*tileGraphics) LoadBitmap(string) (Bitmap, error)           { return nil, nil }
func (*tileGraphics) LoadBitmapTable(string) (BitmapTable, error) { return nil, nil }
func (*tileGraphics) NewBitmap(int, int) (Bitmap, error)          { return nil, nil }
func (graphics *tileGraphics) DrawBitmap(_ Bitmap, x, y int) error {
	graphics.positions = append(graphics.positions, [2]int{x, y})
	return graphics.fail
}
func (*tileGraphics) DrawScaledBitmap(Bitmap, int, int, float32, float32) error { return nil }

func TestTileMapDrawBoundsVisibleWork(t *testing.T) {
	tiles := make([]uint8, 100*100)
	for index := range tiles {
		tiles[index] = 1
	}
	tilemap, err := NewTileMap(TileMapConfig{Columns: 100, Rows: 100, TileWidth: 16, TileHeight: 16, Tiles: tiles})
	if err != nil {
		t.Fatal(err)
	}
	graphics := &tileGraphics{}
	stats, err := tilemap.Draw(graphics, []Bitmap{&tileBitmap{}}, Camera{X: 17, Y: 33, Width: 32, Height: 32})
	if err != nil {
		t.Fatal(err)
	}
	if stats != (TileDrawStats{Visited: 9, Drawn: 9}) {
		t.Fatalf("stats = %+v", stats)
	}
	want := [][2]int{{-1, -1}, {15, -1}, {31, -1}, {-1, 15}, {15, 15}, {31, 15}, {-1, 31}, {15, 31}, {31, 31}}
	if !reflect.DeepEqual(graphics.positions, want) {
		t.Fatalf("positions = %v", graphics.positions)
	}
}

func TestTileMapCollisionAndOwnership(t *testing.T) {
	tiles := []uint8{0, 1, 0, 0}
	tilemap, err := NewTileMap(TileMapConfig{Columns: 2, Rows: 2, TileWidth: 16, TileHeight: 16, Tiles: tiles, Solid: []bool{false, true}})
	if err != nil {
		t.Fatal(err)
	}
	tiles[1] = 0
	if !tilemap.IntersectsSolid(Rect{X: 16, Y: 0, Width: 4, Height: 4}) {
		t.Fatal("solid tile was not detected")
	}
	if tilemap.IntersectsSolid(Rect{X: 12, Y: 0, Width: 4, Height: 4}) {
		t.Fatal("edge contact counted as overlap")
	}
	if !tilemap.IntersectsSolid(Rect{X: 15.5, Y: 0, Width: 1, Height: 4}) {
		t.Fatal("fractional overlap was not detected")
	}
	if tilemap.IntersectsSolid(Rect{X: -0.5, Y: 0, Width: 0.25, Height: 4}) {
		t.Fatal("negative fractional rectangle reached the map")
	}
	if _, ok := tilemap.TileAt(2, 0); ok {
		t.Fatal("out-of-range tile reported present")
	}
}

func TestTileMapValidationAndDrawErrors(t *testing.T) {
	if _, err := NewTileMap(TileMapConfig{Columns: 2, Rows: 2, TileWidth: 16, TileHeight: 16, Tiles: []uint8{1}}); !errors.Is(err, ErrTileMapConfig) {
		t.Fatalf("error = %v", err)
	}
	tilemap, _ := NewTileMap(TileMapConfig{Columns: 1, Rows: 1, TileWidth: 16, TileHeight: 16, Tiles: []uint8{1}})
	if _, err := tilemap.Draw(&tileGraphics{}, nil, Camera{Width: 16, Height: 16}); !errors.Is(err, ErrTileMapBitmap) {
		t.Fatalf("error = %v", err)
	}
	want := errors.New("draw")
	if _, err := tilemap.Draw(&tileGraphics{fail: want}, []Bitmap{&tileBitmap{}}, Camera{Width: 16, Height: 16}); !errors.Is(err, want) || !errors.Is(err, ErrTileMapDraw) {
		t.Fatalf("error = %v", err)
	}
}

func TestCameraClamp(t *testing.T) {
	camera := Camera{X: 900, Y: -4, Width: 400, Height: 240}
	camera.Clamp(640, 480)
	if camera.X != 240 || camera.Y != 0 {
		t.Fatalf("camera = %+v", camera)
	}
}
