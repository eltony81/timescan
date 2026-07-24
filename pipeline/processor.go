package pipeline

import (
	"context"
	"sync"

	"github.com/timescan/timescan/timeseries"
)

// StreamProcessor manages concurrent processing of metric streams over Go channels.
// Adheres to Go concurrency best practices:
// - Producer owns and closes channels.
// - Workers exit cleanly on channel close or context cancellation (no goroutine leaks).
// - Graceful shutdown via sync.WaitGroup.
type StreamProcessor struct {
	engine     *Engine
	numWorkers int
}

// NewStreamProcessor creates a new concurrency-safe StreamProcessor.
func NewStreamProcessor(engine *Engine, numWorkers int) *StreamProcessor {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	return &StreamProcessor{
		engine:     engine,
		numWorkers: numWorkers,
	}
}

// ProcessStream processes incoming DataPoints from an input channel concurrently
// and writes results to an output channel.
func (sp *StreamProcessor) ProcessStream(ctx context.Context, in <-chan timeseries.DataPoint) <-chan Result {
	out := make(chan Result, cap(in))
	var wg sync.WaitGroup

	for i := 0; i < sp.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case dp, ok := <-in:
					if !ok {
						return // Input channel closed, worker terminates gracefully
					}
					res, err := sp.engine.Process(ctx, dp)
					if err == nil {
						select {
						case out <- res:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
	}

	// Monitor routine: closes out channel once all workers finish
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
