package diagnostics

import (
	"errors"
	"testing"
)

func TestCollectorReport(t *testing.T) {
	collector, err := New(5)
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []Sample{
		{FrameMilliseconds: 10, HeapBytes: 100, NativeResources: 4},
		{FrameMilliseconds: 20, HeapBytes: 120, NativeResources: 5},
		{FrameMilliseconds: 30, HeapBytes: 115, NativeResources: 3},
		{FrameMilliseconds: 40, HeapBytes: 140, NativeResources: 6},
		{FrameMilliseconds: 300, HeapBytes: 90, NativeResources: 2},
	} {
		if err := collector.Record(sample); err != nil {
			t.Fatal(err)
		}
	}
	if !collector.Complete() {
		t.Fatal("Complete() = false")
	}
	got := collector.Report()
	if got.Frames != 5 || got.FrameMeanMilliseconds != 80 || got.FrameP50Milliseconds != 30 || got.FrameP95Milliseconds != 300 || got.FrameP99Milliseconds != 300 || got.FrameMaxMilliseconds != 300 {
		t.Fatalf("frame report = %+v", got)
	}
	if got.HeapStartBytes != 100 || got.HeapEndBytes != 90 || got.HeapMaxBytes != 140 || got.HeapGrowthBytes != -10 {
		t.Fatalf("heap report = %+v", got)
	}
	if got.NativeResourcesStart != 4 || got.NativeResourcesEnd != 2 || got.NativeResourcesMinimum != 2 || got.NativeResourcesMaximum != 6 {
		t.Fatalf("resource report = %+v", got)
	}
	if err := collector.Record(Sample{}); !errors.Is(err, ErrComplete) {
		t.Fatalf("Record after completion = %v", err)
	}
	if got := collector.Report(); got.Dropped != 1 {
		t.Fatalf("Dropped = %d, want 1", got.Dropped)
	}
}

func TestCollectorRejectsInvalidLimits(t *testing.T) {
	for _, limit := range []int{0, -1, MaximumFrames + 1} {
		if _, err := New(limit); !errors.Is(err, ErrFrameLimit) {
			t.Errorf("New(%d) error = %v", limit, err)
		}
	}
	var collector *Collector
	if err := collector.Record(Sample{}); !errors.Is(err, ErrFrameLimit) {
		t.Fatalf("nil Record error = %v", err)
	}
	if got := collector.Report(); got != (Report{}) {
		t.Fatalf("nil Report = %+v", got)
	}
}

func TestCollectorReportsPositiveHeapGrowth(t *testing.T) {
	collector, _ := New(2)
	_ = collector.Record(Sample{HeapBytes: 10})
	_ = collector.Record(Sample{HeapBytes: 25})
	if got := collector.Report().HeapGrowthBytes; got != 15 {
		t.Fatalf("HeapGrowthBytes = %d", got)
	}
}

func TestSignedGrowthSaturates(t *testing.T) {
	const max = int64(^uint64(0) >> 1)
	if got := signedGrowth(0, ^uint64(0)); got != max {
		t.Fatalf("positive signedGrowth = %d", got)
	}
	if got := signedGrowth(^uint64(0), 0); got != -max-1 {
		t.Fatalf("negative signedGrowth = %d", got)
	}
}
