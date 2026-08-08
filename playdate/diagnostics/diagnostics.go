// Package diagnostics provides bounded, allocation-free gameplay measurements.
package diagnostics

import "errors"

const (
	// MaximumFrames bounds a collection to twenty minutes at 30 frames/second.
	MaximumFrames = 36_000
	frameBuckets  = 256
)

var (
	// ErrFrameLimit reports an invalid configured collection length.
	ErrFrameLimit = errors.New("diagnostics frame limit must be between 1 and 36000")
	// ErrComplete reports a sample recorded after the configured collection ended.
	ErrComplete = errors.New("diagnostics collection is complete")
)

// Sample is one externally observed gameplay frame. HeapBytes should describe
// live heap rather than cumulative allocations. NativeResources is the number
// of currently owned native resources after the frame.
type Sample struct {
	FrameMilliseconds uint32
	HeapBytes         uint64
	NativeResources   uint32
}

// Report summarizes a completed or partial collection. Frame percentiles use
// nearest-rank selection. HeapGrowth is signed so cleanup is visible.
type Report struct {
	Frames                 uint32
	Dropped                uint32
	FrameMeanMilliseconds  float64
	FrameP50Milliseconds   uint32
	FrameP95Milliseconds   uint32
	FrameP99Milliseconds   uint32
	FrameMaxMilliseconds   uint32
	HeapStartBytes         uint64
	HeapEndBytes           uint64
	HeapMaxBytes           uint64
	HeapGrowthBytes        int64
	NativeResourcesStart   uint32
	NativeResourcesEnd     uint32
	NativeResourcesMinimum uint32
	NativeResourcesMaximum uint32
}

// Collector aggregates a fixed number of samples without allocating while
// frames are recorded. The zero value is invalid; use New.
type Collector struct {
	limit       uint32
	frames      uint32
	dropped     uint32
	frameTotal  uint64
	frameMax    uint32
	buckets     [frameBuckets]uint32
	heapStart   uint64
	heapEnd     uint64
	heapMax     uint64
	resourceMin uint32
	resourceMax uint32
	resourceBeg uint32
	resourceEnd uint32
}

// New creates a collector that accepts exactly frameLimit samples.
func New(frameLimit int) (*Collector, error) {
	if frameLimit < 1 || frameLimit > MaximumFrames {
		return nil, ErrFrameLimit
	}
	return &Collector{limit: uint32(frameLimit)}, nil
}

// Record adds a sample. Once the configured bound is reached it counts the
// rejected sample and returns ErrComplete without changing the report.
func (c *Collector) Record(sample Sample) error {
	if c == nil || c.limit == 0 {
		return ErrFrameLimit
	}
	if c.frames == c.limit {
		c.dropped++
		return ErrComplete
	}
	if c.frames == 0 {
		c.heapStart = sample.HeapBytes
		c.resourceBeg = sample.NativeResources
		c.resourceMin = sample.NativeResources
	}
	c.frames++
	c.frameTotal += uint64(sample.FrameMilliseconds)
	if sample.FrameMilliseconds > c.frameMax {
		c.frameMax = sample.FrameMilliseconds
	}
	bucket := sample.FrameMilliseconds
	if bucket >= frameBuckets {
		bucket = frameBuckets - 1
	}
	c.buckets[bucket]++
	c.heapEnd = sample.HeapBytes
	if sample.HeapBytes > c.heapMax {
		c.heapMax = sample.HeapBytes
	}
	c.resourceEnd = sample.NativeResources
	if sample.NativeResources < c.resourceMin {
		c.resourceMin = sample.NativeResources
	}
	if sample.NativeResources > c.resourceMax {
		c.resourceMax = sample.NativeResources
	}
	return nil
}

// Complete reports whether the configured sample bound has been reached.
func (c *Collector) Complete() bool { return c != nil && c.limit != 0 && c.frames == c.limit }

// Report returns the current aggregate without allocating.
func (c *Collector) Report() Report {
	if c == nil || c.frames == 0 {
		return Report{}
	}
	return Report{
		Frames: c.frames, Dropped: c.dropped,
		FrameMeanMilliseconds: float64(c.frameTotal) / float64(c.frames),
		FrameP50Milliseconds:  c.percentile(50),
		FrameP95Milliseconds:  c.percentile(95),
		FrameP99Milliseconds:  c.percentile(99),
		FrameMaxMilliseconds:  c.frameMax,
		HeapStartBytes:        c.heapStart, HeapEndBytes: c.heapEnd, HeapMaxBytes: c.heapMax,
		HeapGrowthBytes:      signedGrowth(c.heapStart, c.heapEnd),
		NativeResourcesStart: c.resourceBeg, NativeResourcesEnd: c.resourceEnd,
		NativeResourcesMinimum: c.resourceMin, NativeResourcesMaximum: c.resourceMax,
	}
}

func (c *Collector) percentile(percent uint32) uint32 {
	rank := (c.frames*percent + 99) / 100
	var seen uint32
	for milliseconds, count := range c.buckets {
		seen += count
		if seen >= rank {
			if milliseconds == frameBuckets-1 && c.frameMax > frameBuckets-1 {
				return c.frameMax
			}
			return uint32(milliseconds)
		}
	}
	return c.frameMax
}

func signedGrowth(start, end uint64) int64 {
	const max = uint64(^uint64(0) >> 1)
	if end >= start {
		growth := end - start
		if growth > max {
			return int64(max)
		}
		return int64(growth)
	}
	shrink := start - end
	if shrink >= max+1 {
		return -int64(max) - 1
	}
	return -int64(shrink)
}
