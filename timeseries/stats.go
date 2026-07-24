package timeseries

import (
	"math"
	"sort"
)

// Welford represents Welford's online algorithm for computing variance.
type Welford struct {
	count float64
	mean  float64
	m2    float64
}

// NewWelford creates a new Welford stats collector.
func NewWelford() *Welford {
	return &Welford{}
}

// Update adds a new value to the online statistics.
func (w *Welford) Update(val float64) {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return
	}
	w.count++
	delta := val - w.mean
	w.mean += delta / w.count
	delta2 := val - w.mean
	w.m2 += delta * delta2
}

// Mean returns the running mean.
func (w *Welford) Mean() float64 {
	return w.mean
}

// Variance returns the running sample variance.
func (w *Welford) Variance() float64 {
	if w.count < 2 {
		return 0
	}
	return w.m2 / (w.count - 1)
}

// StdDev returns the running standard deviation.
func (w *Welford) StdDev() float64 {
	return math.Sqrt(w.Variance())
}

// Count returns the number of samples seen.
func (w *Welford) Count() int {
	return int(w.count)
}

// Median calculates the median of a float64 slice.
// Returns 0 for empty slices. Does not modify the input slice.
func Median(data []float64) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}
	cp := make([]float64, n)
	copy(cp, data)
	sort.Float64s(cp)
	if n%2 == 0 {
		return (cp[n/2-1] + cp[n/2]) / 2
	}
	return cp[n/2]
}

// IQR calculates the Interquartile Range without mutating the input slice.
func IQR(data []float64) float64 {
	n := len(data)
	if n < 4 {
		return 0
	}
	cp := make([]float64, n)
	copy(cp, data)
	sort.Float64s(cp)

	mid := n / 2
	var q1, q3 float64
	if n%2 == 0 {
		q1 = Median(cp[:mid])
		q3 = Median(cp[mid:])
	} else {
		q1 = Median(cp[:mid])
		q3 = Median(cp[mid+1:])
	}
	return q3 - q1
}

// MAD calculates the Median Absolute Deviation without mutating the input slice.
func MAD(data []float64) float64 {
	n := len(data)
	if n == 0 {
		return 0
	}
	med := Median(data)

	devs := make([]float64, n)
	for i, val := range data {
		devs[i] = math.Abs(val - med)
	}
	return Median(devs)
}
