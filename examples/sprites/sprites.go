// Package sprites exercises the portable P2.1 sprite display-list API.
package sprites

import (
	"errors"

	"github.com/Djunichi/gopdsdk/playdate"
)

const spriteCount = 3

type game struct {
	bitmap  playdate.Bitmap
	sprites [spriteCount]playdate.Sprite
	x       [spriteCount]float32
	frame   int
	closed  bool
}

// New creates the sprite display-list acceptance game.
func New() playdate.Game { return &game{x: [spriteCount]float32{80, 200, 320}} }

func (g *game) Init(context playdate.Context) error {
	bitmap, err := context.NewBitmap(24, 24)
	if err != nil {
		return err
	}
	g.bitmap = bitmap
	if err := bitmap.Fill(playdate.ColorBlack); err != nil {
		return errors.Join(err, g.close())
	}
	for index := range g.sprites {
		sprite, createErr := context.NewSprite()
		if createErr != nil {
			return errors.Join(createErr, g.close())
		}
		g.sprites[index] = sprite
		if setupErr := setupSprite(sprite, bitmap, g.x[index], float32(100+index*42), index); setupErr != nil {
			return errors.Join(setupErr, g.close())
		}
	}
	return nil
}

func setupSprite(sprite playdate.Sprite, bitmap playdate.Bitmap, x, y float32, z int) error {
	if err := sprite.SetBitmap(bitmap); err != nil {
		return err
	}
	if err := sprite.SetPosition(x, y); err != nil {
		return err
	}
	if err := sprite.SetVisible(true); err != nil {
		return err
	}
	if err := sprite.SetZIndex(z); err != nil {
		return err
	}
	return sprite.Add()
}

func (g *game) Update(context playdate.Context) (bool, error) {
	g.frame++
	g.x[0] += context.Input().CrankDelta
	if g.x[0] < 12 {
		g.x[0] = 388
	}
	if g.x[0] > 388 {
		g.x[0] = 12
	}
	if err := g.sprites[0].SetPosition(g.x[0], 100); err != nil {
		return false, err
	}
	if err := g.sprites[1].MoveBy(1, 0); err != nil {
		return false, err
	}
	g.x[1]++
	if g.x[1] > 388 {
		g.x[1] = 12
		if err := g.sprites[1].SetPosition(g.x[1], 142); err != nil {
			return false, err
		}
	}
	if err := g.sprites[2].MoveBy(-1, 0); err != nil {
		return false, err
	}
	g.x[2]--
	if g.x[2] < 12 {
		g.x[2] = 388
		if err := g.sprites[2].SetPosition(g.x[2], 184); err != nil {
			return false, err
		}
	}
	if g.frame%120 == 0 {
		if err := g.sprites[2].Remove(); err != nil {
			return false, err
		}
	}
	if g.frame > 120 && g.frame%120 == 1 {
		if err := g.sprites[2].Add(); err != nil {
			return false, err
		}
	}
	context.Clear()
	context.DrawText("P2.1 sprites: crank + moving list", 12, 8)
	context.UpdateAndDrawSprites()
	return true, nil
}

func (g *game) HandleLifecycle(_ playdate.Context, event playdate.LifecycleEvent) error {
	if event == playdate.LifecycleTerminate {
		return g.close()
	}
	return nil
}

func (g *game) close() error {
	if g.closed {
		return nil
	}
	g.closed = true
	var err error
	for _, sprite := range g.sprites {
		if sprite != nil {
			err = errors.Join(err, sprite.Close())
		}
	}
	if g.bitmap != nil {
		err = errors.Join(err, g.bitmap.Close())
	}
	return err
}
