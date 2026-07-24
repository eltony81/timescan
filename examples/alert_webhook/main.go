package main

import (
	"context"
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

type AlertDispatcher struct {
	lastAlertTime time.Time
	cooldown      time.Duration
}

func (a *AlertDispatcher) Dispatch(dp timeseries.DataPoint, meta anomaly.AnomalyMeta) {
	if time.Since(a.lastAlertTime) < a.cooldown {
		fmt.Printf("[Throttled] Anomaly suppressed to prevent spam (Value: %.2f)\n", dp.Value)
		return
	}

	a.lastAlertTime = time.Now()
	fmt.Printf("[WEBHOOK] POST /alerts -> 'CRITICAL: CPU Spike! Value: %.2f (Score: %.2f)'\n",
		dp.Value, meta.Score)
}

func main() {
	engine := pipeline.NewEngine(pipeline.Config{
		WindowSize: 10,
		Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.0}),
	})

	dispatcher := &AlertDispatcher{
		cooldown: 5 * time.Second,
	}

	fmt.Println("--- Alert Webhook Dispatcher with Debouncing ---")

	ctx := context.Background()
	for i := 0; i < 15; i++ {
		val := 10.0
		if i >= 5 && i <= 10 {
			val = 100.0
		} else {
			val = 10.0 + float64(i%2)
		}

		dp := timeseries.DataPoint{
			Timestamp: time.Now(),
			Value:     val,
		}

		res, _ := engine.Process(ctx, dp)
		if res.IsAnomaly {
			dispatcher.Dispatch(dp, res.AnomalyMeta)
		} else {
			fmt.Printf("[OK] Value: %.2f\n", val)
		}

		time.Sleep(1 * time.Second)
	}
}
