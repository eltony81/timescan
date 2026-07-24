package pipeline

import (
	"context"
	"sync"
	"testing"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/timeseries"
)

func TestPipelineEngine(t *testing.T) {
	engine := NewEngine(Config{
		WindowSize: 10,
		Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.0}),
	})

	var wg sync.WaitGroup
	ctx := context.Background()
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			_, _ = engine.Process(ctx, timeseries.DataPoint{Value: val})
		}(float64(i))
	}
	wg.Wait()
}
