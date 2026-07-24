package decomposition

import (
	"testing"
	"time"

	"github.com/timescan/timescan/timeseries"
)

func TestDecomposeAdditive(t *testing.T) {
	series := timeseries.Series{Points: make([]timeseries.DataPoint, 20)}
	for i := 0; i < 20; i++ {
		series.Points[i] = timeseries.DataPoint{
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Value:     float64(i + (i % 4)),
		}
	}

	decomp := DecomposeAdditive(series, 4)
	if len(decomp.Trend.Points) != 20 {
		t.Errorf("expected 20 points in trend, got %d", len(decomp.Trend.Points))
	}
}

func FuzzDecomposeAdditive(f *testing.F) {
	f.Add(4, 10)
	f.Add(0, 5)
	f.Add(-1, 20)
	f.Fuzz(func(t *testing.T, period int, numPoints int) {
		if numPoints < 0 || numPoints > 200 {
			return
		}
		series := timeseries.Series{Points: make([]timeseries.DataPoint, numPoints)}
		for i := 0; i < numPoints; i++ {
			series.Points[i] = timeseries.DataPoint{Value: float64(i)}
		}
		_ = DecomposeAdditive(series, period)
	})
}
