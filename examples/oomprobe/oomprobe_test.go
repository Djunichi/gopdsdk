package oomprobe

import (
	"testing"

	"github.com/Djunichi/gopdsdk/playdate"
)

type testContext struct{}

func (testContext) Clear()                          {}
func (testContext) DrawText(string, int, int)       {}
func (testContext) CurrentTimeMilliseconds() uint32 { return 0 }
func (testContext) Input() playdate.Input           { return playdate.Input{} }

func TestUpdateRetainsEveryAllocatedBlock(t *testing.T) {
	game := &game{}
	const frames = 4
	for range frames {
		refresh, err := game.Update(testContext{})
		if err != nil || !refresh {
			t.Fatalf("Update() = %v, %v; want true, nil", refresh, err)
		}
	}
	if game.frame != frames {
		t.Fatalf("frame = %d, want %d", game.frame, frames)
	}
	for index := range frames {
		if len(game.blocks[index]) != blockSize {
			t.Fatalf("blocks[%d] length = %d, want %d", index, len(game.blocks[index]), blockSize)
		}
	}
}
