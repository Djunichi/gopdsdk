// Package gcstress provides a bounded allocation workload for device GC acceptance.
package gcstress

import (
	"runtime"

	"github.com/Djunichi/gopdsdk/playdate"
)

const (
	blockSize      = 2 * 1024
	retainedBlocks = 16
	collectionRate = 16
	heartbeatRate  = 30
)

type game struct {
	blocks    [retainedBlocks][]byte
	frame     uint32
	checksum  byte
	heartbeat bool
}

// New creates a game that continuously replaces a bounded set of heap blocks.
func New() playdate.Game { return &game{} }

func (g *game) Init(playdate.Context) error { return nil }

func (g *game) Update(context playdate.Context) (bool, error) {
	block := make([]byte, blockSize)
	for index := 0; index < len(block); index += 256 {
		value := byte(g.frame) ^ byte(index>>8)
		block[index] = value
		g.checksum ^= value
	}
	g.blocks[g.frame%retainedBlocks] = block
	g.frame++
	if g.frame%collectionRate == 0 {
		runtime.GC()
	}
	if g.frame%heartbeatRate == 0 {
		g.heartbeat = !g.heartbeat
	}

	context.Clear()
	if g.heartbeat {
		context.DrawText("GC stress: running +", 16, 16)
	} else {
		context.DrawText("GC stress: running -", 16, 16)
	}
	return true, nil
}
