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
	heapSize       = 256 * 1024
	maxLiveHeap    = heapSize * 3 / 4
	frameBudgetMS  = 33
)

type game struct {
	blocks       [retainedBlocks][]byte
	frame        uint32
	checksum     byte
	heartbeat    bool
	bounded      bool
	failed       bool
	timingFailed bool
	maxUpdateMS  uint32
	maxGCMS      uint32
	stats        runtime.MemStats
	collect      func()
	readStats    func(*runtime.MemStats)
}

// New creates a game that continuously replaces a bounded set of heap blocks.
func New() playdate.Game { return newGame(runtime.GC, runtime.ReadMemStats) }

func newGame(collect func(), readStats func(*runtime.MemStats)) *game {
	return &game{collect: collect, readStats: readStats}
}

func (g *game) Init(playdate.Context) error { return nil }

func (g *game) Update(context playdate.Context) (bool, error) {
	frameStart := context.CurrentTimeMilliseconds()
	block := make([]byte, blockSize)
	for index := 0; index < len(block); index += 256 {
		value := byte(g.frame) ^ byte(index>>8)
		block[index] = value
		g.checksum ^= value
	}
	g.blocks[g.frame%retainedBlocks] = block
	g.frame++
	if g.frame%collectionRate == 0 {
		gcStart := context.CurrentTimeMilliseconds()
		g.collect()
		gcDuration := context.CurrentTimeMilliseconds() - gcStart
		if gcDuration > g.maxGCMS {
			g.maxGCMS = gcDuration
		}
		g.readStats(&g.stats)
		g.failed = g.failed || g.stats.HeapAlloc > maxLiveHeap
		g.bounded = !g.failed && g.stats.TotalAlloc > heapSize && g.stats.Frees > 0
	}
	if g.frame%heartbeatRate == 0 {
		g.heartbeat = !g.heartbeat
	}

	context.Clear()
	if g.failed {
		context.DrawText("GC stress: FAILED", 16, 16)
	} else if g.timingFailed {
		context.DrawText("GC stress: FRAME FAIL", 16, 16)
	} else if g.bounded && g.heartbeat {
		context.DrawText("GC stress: timing ok +", 16, 16)
	} else if g.bounded {
		context.DrawText("GC stress: timing ok -", 16, 16)
	} else if g.heartbeat {
		context.DrawText("GC stress: warming +", 16, 16)
	} else {
		context.DrawText("GC stress: warming -", 16, 16)
	}
	updateDuration := context.CurrentTimeMilliseconds() - frameStart
	if updateDuration > g.maxUpdateMS {
		g.maxUpdateMS = updateDuration
	}
	g.timingFailed = g.timingFailed || updateDuration > frameBudgetMS || g.maxGCMS > frameBudgetMS
	return true, nil
}
