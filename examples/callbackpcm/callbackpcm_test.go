package callbackpcm

import "testing"

func TestRenderProducesAndCanStarve(t *testing.T) {
	g := &game{frequency: 220}
	buffer := make([]int16, 32)
	right := make([]int16, len(buffer))
	if got := g.render(buffer, right); got != len(buffer) || g.phase == 0 || buffer[1] == 0 || right[1] != 0 {
		t.Fatalf("render = %d, phase %v", got, g.phase)
	}
	g.right = true
	g.render(buffer, right)
	if buffer[1] != 0 || right[1] == 0 {
		t.Fatalf("right routing left=%d right=%d", buffer[1], right[1])
	}
	g.starve = true
	if got := g.render(buffer, right); got != 0 {
		t.Fatalf("starved render = %d", got)
	}
}
