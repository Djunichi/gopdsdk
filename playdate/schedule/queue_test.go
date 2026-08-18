package schedule

import (
	"reflect"
	"testing"
)

func TestQueueBoundedFIFOAndClear(t *testing.T) {
	queue, err := NewQueue[int](2)
	if err != nil {
		t.Fatal(err)
	}
	if queue.Capacity() != 2 || !queue.TrySend(1) || !queue.TrySend(2) || queue.TrySend(3) {
		t.Fatalf("unexpected queue state: len %d", queue.Len())
	}
	first, ready := queue.TryReceive()
	if !ready || first != 1 || !queue.TrySend(3) {
		t.Fatalf("first = %d, ready %v", first, ready)
	}
	second, _ := queue.TryReceive()
	third, _ := queue.TryReceive()
	if second != 2 || third != 3 {
		t.Fatalf("received %d, %d", second, third)
	}
	if _, ready = queue.TryReceive(); ready {
		t.Fatal("empty queue reported ready")
	}
	queue.TrySend(4)
	queue.Clear()
	if queue.Len() != 0 {
		t.Fatalf("len after clear = %d", queue.Len())
	}
}

func TestPollerRoundRobinAndNothingReady(t *testing.T) {
	first, _ := NewQueue[int](3)
	second, _ := NewQueue[int](3)
	poller, err := NewPoller(first, second)
	if err != nil {
		t.Fatal(err)
	}
	first.TrySend(10)
	first.TrySend(11)
	second.TrySend(20)
	second.TrySend(21)
	var values []int
	var indexes []int
	for range 4 {
		value, index, ready := poller.TryReceive()
		if !ready {
			t.Fatal("poller unexpectedly empty")
		}
		values = append(values, value)
		indexes = append(indexes, index)
	}
	if !reflect.DeepEqual(values, []int{10, 20, 11, 21}) || !reflect.DeepEqual(indexes, []int{0, 1, 0, 1}) {
		t.Fatalf("values %v, indexes %v", values, indexes)
	}
	value, index, ready := poller.TryReceive()
	if ready || index != -1 || value != 0 {
		t.Fatalf("empty poll = %d, %d, %v", value, index, ready)
	}
}
