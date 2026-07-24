package timeseries

import (
	"sync"
)

// RingBufferWindow is a thread-safe, fixed-capacity sliding window for real-time streams
// with O(1) updates and zero allocation path if reusing snapshot slices.
type RingBufferWindow struct {
	mu       sync.RWMutex
	capacity int
	buffer   []DataPoint
	head     int
	count    int
}

// NewRingBufferWindow creates a new ring buffer with a fixed capacity.
// If capacity <= 0, it defaults to 1 to prevent division-by-zero panics.
func NewRingBufferWindow(capacity int) *RingBufferWindow {
	if capacity <= 0 {
		capacity = 1
	}
	return &RingBufferWindow{
		capacity: capacity,
		buffer:   make([]DataPoint, capacity),
	}
}

// Push adds a new data point to the buffer, overwriting the oldest if full.
func (r *RingBufferWindow) Push(dp DataPoint) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buffer[r.head] = dp
	r.head = (r.head + 1) % r.capacity
	if r.count < r.capacity {
		r.count++
	}
}

// Snapshot returns a copy of the current window to avoid race conditions during analysis.
// To achieve zero allocations, pass a pre-allocated destination slice.
func (r *RingBufferWindow) Snapshot(dst []DataPoint) []DataPoint {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if cap(dst) < r.count {
		dst = make([]DataPoint, r.count)
	} else {
		dst = dst[:r.count]
	}

	if r.count == 0 {
		return dst
	}

	if r.count < r.capacity {
		copy(dst, r.buffer[:r.count])
	} else {
		// Buffer is full, read from head to end, then 0 to head
		n := copy(dst, r.buffer[r.head:])
		copy(dst[n:], r.buffer[:r.head])
	}

	return dst
}

// Count returns the current number of items in the buffer.
func (r *RingBufferWindow) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.count
}
