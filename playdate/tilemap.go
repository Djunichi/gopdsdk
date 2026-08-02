// Package playdate defines the portable API shared by Playdate games and native adapters.
package playdate

import (
	"errors"
)

// Camera identifies a world-space viewport in integer pixels.
type Camera struct {
	X, Y          int
	Width, Height int
}

// Clamp keeps the camera viewport inside a world of the given size.
func (camera *Camera) Clamp(worldWidth, worldHeight int) {
	if camera == nil {
		return
	}
	maxX, maxY := worldWidth-camera.Width, worldHeight-camera.Height
	if maxX < 0 {
		maxX = 0
	}
	if maxY < 0 {
		maxY = 0
	}
	if camera.X < 0 {
		camera.X = 0
	} else if camera.X > maxX {
		camera.X = maxX
	}
	if camera.Y < 0 {
		camera.Y = 0
	} else if camera.Y > maxY {
		camera.Y = maxY
	}
}

// TileMapConfig describes a row-major tile layer. Tile zero is empty; other
// values select the bitmap at value-1. Solid is indexed by the tile value.
type TileMapConfig struct {
	Columns, Rows         int
	TileWidth, TileHeight int
	Tiles                 []uint8
	Solid                 []bool
}

// TileMap is an immutable-size tile layer with copied tile and collision data.
type TileMap struct {
	columns, rows         int
	tileWidth, tileHeight int
	tiles                 []uint8
	solid                 []bool
}

// TileDrawStats reports the bounded visible work performed by Draw.
type TileDrawStats struct {
	Visited, Drawn int
}

// NewTileMap validates and copies a tile layer.
func NewTileMap(config TileMapConfig) (*TileMap, error) {
	if config.Columns <= 0 || config.Rows <= 0 || config.TileWidth <= 0 || config.TileHeight <= 0 ||
		config.Columns > int(^uint(0)>>1)/config.Rows || config.Columns > int(^uint(0)>>1)/config.TileWidth ||
		config.Rows > int(^uint(0)>>1)/config.TileHeight || len(config.Tiles) != config.Columns*config.Rows {
		return nil, ErrTileMapConfig
	}
	return &TileMap{
		columns: config.Columns, rows: config.Rows, tileWidth: config.TileWidth, tileHeight: config.TileHeight,
		tiles: append([]uint8(nil), config.Tiles...), solid: append([]bool(nil), config.Solid...),
	}, nil
}

// WorldSize returns the layer dimensions in pixels.
func (tilemap *TileMap) WorldSize() (width, height int) {
	if tilemap == nil {
		return 0, 0
	}
	return tilemap.columns * tilemap.tileWidth, tilemap.rows * tilemap.tileHeight
}

// TileAt returns the tile value and whether the coordinates are in the layer.
func (tilemap *TileMap) TileAt(column, row int) (uint8, bool) {
	if tilemap == nil || column < 0 || column >= tilemap.columns || row < 0 || row >= tilemap.rows {
		return 0, false
	}
	return tilemap.tiles[row*tilemap.columns+column], true
}

// Draw renders only tiles intersecting camera. Empty tiles perform no draw.
func (tilemap *TileMap) Draw(graphics Graphics, bitmaps []Bitmap, camera Camera) (TileDrawStats, error) {
	if tilemap == nil || graphics == nil || camera.Width <= 0 || camera.Height <= 0 {
		return TileDrawStats{}, ErrTileMapDraw
	}
	firstColumn, firstRow := floorDiv(camera.X, tilemap.tileWidth), floorDiv(camera.Y, tilemap.tileHeight)
	lastColumn := floorDiv(camera.X+camera.Width-1, tilemap.tileWidth)
	lastRow := floorDiv(camera.Y+camera.Height-1, tilemap.tileHeight)
	firstColumn, firstRow = max(firstColumn, 0), max(firstRow, 0)
	lastColumn, lastRow = min(lastColumn, tilemap.columns-1), min(lastRow, tilemap.rows-1)
	var stats TileDrawStats
	for row := firstRow; row <= lastRow; row++ {
		for column := firstColumn; column <= lastColumn; column++ {
			stats.Visited++
			value := tilemap.tiles[row*tilemap.columns+column]
			if value == 0 {
				continue
			}
			index := int(value) - 1
			if index >= len(bitmaps) || bitmaps[index] == nil {
				return stats, ErrTileMapBitmap
			}
			if err := graphics.DrawBitmap(bitmaps[index], column*tilemap.tileWidth-camera.X, row*tilemap.tileHeight-camera.Y); err != nil {
				return stats, errors.Join(ErrTileMapDraw, err)
			}
			stats.Drawn++
		}
	}
	return stats, nil
}

// IntersectsSolid reports whether an axis-aligned world rectangle overlaps a
// solid tile. Touching an edge alone is not an overlap.
func (tilemap *TileMap) IntersectsSolid(rect Rect) bool {
	if tilemap == nil || rect.Width <= 0 || rect.Height <= 0 {
		return false
	}
	firstColumn := floorFloat32(rect.X / float32(tilemap.tileWidth))
	firstRow := floorFloat32(rect.Y / float32(tilemap.tileHeight))
	lastColumn := floorFloat32((rect.X + rect.Width) / float32(tilemap.tileWidth))
	lastRow := floorFloat32((rect.Y + rect.Height) / float32(tilemap.tileHeight))
	for row := max(firstRow, 0); row <= min(lastRow, tilemap.rows-1); row++ {
		for column := max(firstColumn, 0); column <= min(lastColumn, tilemap.columns-1); column++ {
			value := tilemap.tiles[row*tilemap.columns+column]
			tileLeft, tileTop := float32(column*tilemap.tileWidth), float32(row*tilemap.tileHeight)
			overlaps := rect.X < tileLeft+float32(tilemap.tileWidth) && rect.X+rect.Width > tileLeft &&
				rect.Y < tileTop+float32(tilemap.tileHeight) && rect.Y+rect.Height > tileTop
			if overlaps && int(value) < len(tilemap.solid) && tilemap.solid[value] {
				return true
			}
		}
	}
	return false
}

func floorFloat32(value float32) int {
	integer := int(value)
	if value < 0 && float32(integer) != value {
		return integer - 1
	}
	return integer
}

func floorDiv(value, divisor int) int {
	if value >= 0 {
		return value / divisor
	}
	return -((-value + divisor - 1) / divisor)
}
