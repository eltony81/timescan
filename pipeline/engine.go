package pipeline

import (
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

// Process evaluates a single DataPoint through the pipeline and returns the Result.
func (e *Engine) Process(dp timeseries.DataPoint) Result {
	// Snapshot the historical window context BEFORE pushing the new point
	snap := e.window.Snapshot(nil)
	ctx := anomaly.WindowContext{Window: snap}

	// Evaluate the new incoming point against the historical context
	isAnomaly, meta := e.config.Detector.IsAnomaly(dp, ctx)

	// Now push the new point into the ring buffer window
	e.window.Push(dp)

	return Result{
		IsAnomaly:     isAnomaly,
		AnomalyMeta:   meta,
		WindowContext: ctx,
	}
}
