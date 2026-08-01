package gcstress

import "testing"

type testContext struct{}

func (testContext) Clear()                    {}
func (testContext) DrawText(string, int, int) {}

func TestUpdateKeepsBoundedLiveWindow(t *testing.T) {
	game := &game{}
	context := testContext{}
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
	game := &game{}
	context := testContext{}
	for range heartbeatRate {
		if _, err := game.Update(context); err != nil {
			t.Fatal(err)
		}
	}
	if !game.heartbeat {
		t.Fatal("heartbeat did not change after heartbeatRate frames")
	}
}
