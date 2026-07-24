package pipeline

import (
	"context"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/timeseries"
	"github.com/timescan/timescan/vector"
)

type Config struct {
	WindowSize  int
	Detector    anomaly.Detector
	VectorStore vector.Store
}

// Engine acts as the pipeline coordinator for a single stream of metrics.
type Engine struct {
	config Config
	window *timeseries.RingBufferWindow
}

func NewEngine(config Config) *Engine {
	return &Engine{
		config: config,
		window: timeseries.NewRingBufferWindow(config.WindowSize),
	}
}

// Result holds the pipeline evaluation output for a single point.
type Result struct {
	IsAnomaly     bool
	AnomalyMeta   anomaly.AnomalyMeta
	WindowContext anomaly.WindowContext
}

// Process evaluates a single DataPoint through the pipeline with context support.
// Returns ctx.Err() if the context was cancelled before evaluation.
func (e *Engine) Process(ctx context.Context, dp timeseries.DataPoint) (Result, error) {
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
	}

	snap := e.window.Snapshot(nil)
	wCtx := anomaly.WindowContext{Window: snap}

	isAnomaly, meta := e.config.Detector.IsAnomaly(dp, wCtx)
	e.window.Push(dp)

	return Result{
		IsAnomaly:     isAnomaly,
		AnomalyMeta:   meta,
		WindowContext: wCtx,
	}, nil
}
