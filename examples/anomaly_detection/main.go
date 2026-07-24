package main

import (
	"context"
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

func main() {
	engine := pipeline.NewEngine(pipeline.Config{
		WindowSize: 50,
		Detector: anomaly.NewZScore(anomaly.ZScoreConfig{
			Threshold: 2.0,
		}),
		VectorStore: nil,
	})

	fmt.Println("Starting Anomaly Detection Simulation...")
	fmt.Println("----------------------------------------")

	ctx := context.Background()
	now := time.Now()
	for i := 0; i < 20; i++ {
		val := 100.0
		if i == 15 {
			val = 250.0
		} else if i%2 == 0 {
			val = 105.0
		}

		dp := timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Value:     val,
			Tags:      map[string]string{"host": "server-01", "metric": "cpu_usage"},
		}

		result, _ := engine.Process(ctx, dp)
		timeStr := dp.Timestamp.Format("15:04:05")

		if result.IsAnomaly {
			fmt.Printf("[ALERT] Anomaly Detected at %s!\n", timeStr)
			fmt.Printf("        Value: %.2f (The system expected around ~%.2f)\n",
				dp.Value, result.AnomalyMeta.Expected)
		} else {
			fmt.Printf("[OK]    Value: %.2f at %s\n", dp.Value, timeStr)
		}
	}
}
