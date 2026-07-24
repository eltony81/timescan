package timeseries

import (
	"testing"
	"time"
)

func TestRingBufferWindow(t *testing.T) {
	rb := NewRingBufferWindow(3)
	
	if rb.Count() != 0 {
		t.Fatalf("expected 0, got %d", rb.Count())
	}

	dp1 := DataPoint{Value: 1}
	dp2 := DataPoint{Value: 2}
	dp3 := DataPoint{Value: 3}
	dp4 := DataPoint{Value: 4}

	rb.Push(dp1)
	rb.Push(dp2)
	
	snap := rb.Snapshot(nil)
	if len(snap) != 2 {
		t.Fatalf("expected 2, got %d", len(snap))
	}
	if snap[0].Value != 1 || snap[1].Value != 2 {
		t.Fatalf("unexpected snapshot: %v", snap)
	}

	rb.Push(dp3)
	rb.Push(dp4)

	snap2 := make([]DataPoint, 0, 10)
	snap2 = rb.Snapshot(snap2)
	
	if len(snap2) != 3 {
		t.Fatalf("expected 3, got %d", len(snap2))
	}
	
	if snap2[0].Value != 2 || snap2[1].Value != 3 || snap2[2].Value != 4 {
		t.Fatalf("unexpected snapshot after overflow: %v", snap2)
	}
}

func BenchmarkRingBufferPush(b *testing.B) {
	rb := NewRingBufferWindow(100)
	dp := DataPoint{Timestamp: time.Now(), Value: 42.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Push(dp)
	}
}

func BenchmarkRingBufferSnapshot_Alloc(b *testing.B) {
	rb := NewRingBufferWindow(100)
	for i := 0; i < 100; i++ {
		rb.Push(DataPoint{Value: float64(i)})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rb.Snapshot(nil)
	}
}

func BenchmarkRingBufferSnapshot_ZeroAlloc(b *testing.B) {
	rb := NewRingBufferWindow(100)
	for i := 0; i < 100; i++ {
		rb.Push(DataPoint{Value: float64(i)})
	}

	dst := make([]DataPoint, 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst = rb.Snapshot(dst)
	}
}

func FuzzRingBuffer(f *testing.F) {
	f.Add(10, 5)
	f.Add(0, 0)
	f.Add(-5, 10)
	f.Fuzz(func(t *testing.T, cap int, pushes int) {
		rb := NewRingBufferWindow(cap)
		for i := 0; i < pushes%100; i++ {
			rb.Push(DataPoint{Value: float64(i)})
		}
		snap := rb.Snapshot(nil)
		if len(snap) > rb.capacity {
			t.Fatalf("snapshot length %d > capacity %d", len(snap), rb.capacity)
		}
	})
}
