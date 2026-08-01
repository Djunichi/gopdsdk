package gcstress

import (
	"runtime"
	"testing"
)

type testContext struct{ milliseconds uint32 }

func (testContext) Clear()                    {}
func (testContext) DrawText(string, int, int) {}
func (context *testContext) CurrentTimeMilliseconds() uint32 {
	context.milliseconds++
	return context.milliseconds
}

type timingContext struct {
	milliseconds uint32
	step         uint32
}

func (*timingContext) Clear()                    {}
func (*timingContext) DrawText(string, int, int) {}
func (context *timingContext) CurrentTimeMilliseconds() uint32 {
	current := context.milliseconds
	context.milliseconds += context.step
	return current
}

func TestUpdateKeepsBoundedLiveWindow(t *testing.T) {
	game := newGame(func() {}, func(*runtime.MemStats) {})
	context := &testContext{}
	const frames = retainedBlocks * 8
	for range frames {
		refresh, err := game.Update(context)
		if err != nil || !refresh {
			t.Fatalf("Update() = %v, %v; want true, nil", refresh, err)
		}
	}
	if game.frame != frames {
		t.Fatalf("frame = %d, want %d", game.frame, frames)
	}
	for index, block := range game.blocks {
		if len(block) != blockSize {
			t.Fatalf("blocks[%d] length = %d, want %d", index, len(block), blockSize)
		}
	}
}

func TestHeartbeatChangesAtFixedInterval(t *testing.T) {
	game := newGame(func() {}, func(*runtime.MemStats) {})
	context := &testContext{}
	for range heartbeatRate {
		if _, err := game.Update(context); err != nil {
			t.Fatal(err)
		}
	}
	if !game.heartbeat {
		t.Fatal("heartbeat did not change after heartbeatRate frames")
	}
}

func TestTelemetryReportsBoundedAfterHeapReuse(t *testing.T) {
	collections := 0
	game := newGame(func() { collections++ }, func(stats *runtime.MemStats) {
		stats.HeapAlloc = retainedBlocks * blockSize
		stats.TotalAlloc = heapSize + blockSize
		stats.Frees = 100
	})
	for range collectionRate {
		if _, err := game.Update(&testContext{}); err != nil {
			t.Fatal(err)
		}
	}
	if collections != 1 || !game.bounded || game.failed {
		t.Fatalf("collections/bounded/failed = %d/%v/%v, want 1/true/false", collections, game.bounded, game.failed)
	}
}

func TestTelemetryFailsWhenLiveHeapExceedsBudget(t *testing.T) {
	game := newGame(func() {}, func(stats *runtime.MemStats) {
		stats.HeapAlloc = maxLiveHeap + 1
	})
	for range collectionRate {
		if _, err := game.Update(&testContext{}); err != nil {
			t.Fatal(err)
		}
	}
	if !game.failed || game.bounded {
		t.Fatalf("failed/bounded = %v/%v, want true/false", game.failed, game.bounded)
	}
}

func TestTimingFailsWhenUpdateExceedsFrameBudget(t *testing.T) {
	game := newGame(func() {}, func(*runtime.MemStats) {})
	context := &timingContext{step: frameBudgetMS + 1}
	if _, err := game.Update(context); err != nil {
		t.Fatal(err)
	}
	if !game.timingFailed || game.maxUpdateMS <= frameBudgetMS {
		t.Fatalf("timingFailed/maxUpdateMS = %v/%d, want true/>%d", game.timingFailed, game.maxUpdateMS, frameBudgetMS)
	}
}

func TestTimingHandlesMillisecondCounterWrap(t *testing.T) {
	game := newGame(func() {}, func(*runtime.MemStats) {})
	context := &timingContext{milliseconds: ^uint32(0) - 1, step: 2}
	if _, err := game.Update(context); err != nil {
		t.Fatal(err)
	}
	if game.maxUpdateMS != 2 || game.timingFailed {
		t.Fatalf("maxUpdateMS/timingFailed = %d/%v, want 2/false", game.maxUpdateMS, game.timingFailed)
	}
}
