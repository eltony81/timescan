package anomaly

import (
	"testing"
	"time"

	"github.com/timescan/timescan/timeseries"
)

func TestDetectors(t *testing.T) {
	// A series with baseline variation so stdDev and MAD are non-zero
	series := timeseries.Series{
		Points: []timeseries.DataPoint{
			{Value: 10}, {Value: 12}, {Value: 10}, {Value: 11}, {Value: 100},
		},
	}

	t.Run("ZScore", func(t *testing.T) {
		d := NewZScore(ZScoreConfig{Threshold: 1.5})
		res := d.Detect(series)
		if len(res) == 0 {
			t.Error("expected anomaly detected by ZScore")
		}
	})

	t.Run("EWMA", func(t *testing.T) {
		d := NewEWMA(EWMAConfig{Alpha: 0.5, Threshold: 1.5})
		res := d.Detect(series)
		if len(res) == 0 {
			t.Error("expected anomaly detected by EWMA")
		}
	})

	t.Run("MAD", func(t *testing.T) {
		d := NewMAD(MADConfig{Threshold: 1.5})
		res := d.Detect(series)
		if len(res) == 0 {
			t.Error("expected anomaly detected by MAD")
		}
	})
}

func FuzzDetectors(f *testing.F) {
	f.Add(10.0, 12.0, 100.0, 2.0)
	f.Fuzz(func(t *testing.T, v1, v2, v3, threshold float64) {
		series := timeseries.Series{
			Points: []timeseries.DataPoint{
				{Value: v1}, {Value: v2}, {Value: v3},
			},
		}

		z := NewZScore(ZScoreConfig{Threshold: threshold})
		_ = z.Detect(series)

		ew := NewEWMA(EWMAConfig{Alpha: 0.2, Threshold: threshold})
		_ = ew.Detect(series)

		mad := NewMAD(MADConfig{Threshold: threshold})
		_ = mad.Detect(series)

		ctx := WindowContext{Window: series.Points}
		dp := timeseries.DataPoint{Timestamp: time.Now(), Value: v3}
		_, _ = z.IsAnomaly(dp, ctx)
		_, _ = ew.IsAnomaly(dp, ctx)
		_, _ = mad.IsAnomaly(dp, ctx)
	})
}
