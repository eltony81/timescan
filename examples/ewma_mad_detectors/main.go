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
	engineEWMA := pipeline.NewEngine(pipeline.Config{
		WindowSize: 50,
		Detector: anomaly.NewEWMA(anomaly.EWMAConfig{
			Alpha:     0.15,
			Threshold: 3.0,
		}),
	})

	engineMAD := pipeline.NewEngine(pipeline.Config{
		WindowSize: 50,
		Detector: anomaly.NewMAD(anomaly.MADConfig{
			Threshold: 3.5,
		}),
	})

	fmt.Println("--- Advanced Detectors (EWMA & MAD) ---")

	ctx := context.Background()
	now := time.Now()
	for i := 0; i < 30; i++ {
		val := 10.0 + float64(i%3)

		if i == 10 {
			val = 1000.0
		} else if i == 25 {
			val = 40.0
		}

		dp := timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Value:     val,
		}

		resEWMA, _ := engineEWMA.Process(ctx, dp)
		resMAD, _ := engineMAD.Process(ctx, dp)

		if resEWMA.IsAnomaly || resMAD.IsAnomaly {
			fmt.Printf("Step %d (Value: %.2f):\n", i, val)
			if resEWMA.IsAnomaly {
				fmt.Printf(" - [EWMA] Detected Anomaly! (Score: %.2f)\n", resEWMA.AnomalyMeta.Score)
			}
			if resMAD.IsAnomaly {
				fmt.Printf(" - [MAD]  Detected Anomaly! (Score: %.2f)\n", resMAD.AnomalyMeta.Score)
			}
			fmt.Println()
		}
	}
	fmt.Println("Notice how different algorithms react differently to extreme outliers!")
}
