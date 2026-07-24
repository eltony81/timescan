package vector

import (
	"testing"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/timeseries"
)

func TestPAAEncode(t *testing.T) {
	ctx := anomaly.WindowContext{
		Window: []timeseries.DataPoint{
			{Value: 1}, {Value: 2}, {Value: 3}, {Value: 4},
		},
	}

	vec := PAAEncode(ctx, 2)
	if len(vec) != 2 {
		t.Fatalf("expected length 2, got %d", len(vec))
	}

	if res := PAAEncode(ctx, 0); res != nil {
		t.Errorf("expected nil for 0 dimensions, got %v", res)
	}
}

func FuzzPAAEncode(f *testing.F) {
	f.Add(2, 5)
	f.Add(0, 0)
	f.Add(-1, 10)
	f.Fuzz(func(t *testing.T, dims int, windowLen int) {
		if windowLen < 0 || windowLen > 500 {
			return
		}
		pts := make([]timeseries.DataPoint, windowLen)
		for i := 0; i < windowLen; i++ {
			pts[i] = timeseries.DataPoint{Value: float64(i)}
		}
		ctx := anomaly.WindowContext{Window: pts}
		_ = PAAEncode(ctx, dims)
	})
}
