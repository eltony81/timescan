package vector

import (
	"github.com/timescan/timescan/anomaly"
)

// PAAEncode computes the Piecewise Aggregate Approximation of the window.
// It resamples the time series into the specified number of dimensions.
// Returns nil if dimensions <= 0 or if the window is empty.
func PAAEncode(ctx anomaly.WindowContext, dimensions int) []float32 {
	if dimensions <= 0 {
		return nil
	}

	n := len(ctx.Window)
	if n == 0 {
		return make([]float32, dimensions)
	}

	res := make([]float32, dimensions)

	// If window size is less than dimensions, pad with available values
	if n < dimensions {
		for i := 0; i < n; i++ {
			res[i] = float32(ctx.Window[i].Value)
		}
		return res
	}

	segmentSize := float64(n) / float64(dimensions)

	for i := 0; i < dimensions; i++ {
		start := float64(i) * segmentSize
		end := start + segmentSize

		sum := 0.0
		for j := int(start); j < int(end) && j < n; j++ {
			sum += ctx.Window[j].Value
		}

		count := int(end) - int(start)
		if count > 0 {
			res[i] = float32(sum / float64(count))
		}
	}

	return res
}
