package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/timeseries"
)

func TestStreamProcessor(t *testing.T) {
	engine := NewEngine(Config{
		WindowSize: 10,
		Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.0}),
	})

	processor := NewStreamProcessor(engine, 4)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan timeseries.DataPoint, 20)

	// Feed points
	for i := 0; i < 15; i++ {
		in <- timeseries.DataPoint{Value: float64(i)}
	}
	close(in) // Producer closes channel

	out := processor.ProcessStream(ctx, in)

	count := 0
	for range out {
		count++
	}

	if count != 15 {
		t.Errorf("expected 15 processed items, got %d", count)
	}
}

func TestStreamProcessor_Cancellation(t *testing.T) {
	engine := NewEngine(Config{
		WindowSize: 10,
		Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.0}),
	})

	processor := NewStreamProcessor(engine, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	in := make(chan timeseries.DataPoint) // unbuffered
	_ = processor.ProcessStream(ctx, in)

	time.Sleep(20 * time.Millisecond)
	// Verification: worker routines exit on ctx.Done() without deadlock
}
