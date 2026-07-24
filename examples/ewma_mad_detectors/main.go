package main

import (
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

func main() {
	// 1. CONFIGURE THE EWMA ENGINE
	// EWMA (Exponentially Weighted Moving Average) is a detector that gives more
	// importance (weight) to very recent data. It's great for tracking metrics 
	// that drift slowly over time but suddenly change abruptly.
	engineEWMA := pipeline.NewEngine(pipeline.Config{
		WindowSize: 50,
		Detector: anomaly.NewEWMA(anomaly.EWMAConfig{
			Alpha:     0.15, // Alpha controls how much we care about recent vs old data
			Threshold: 3.0,  // Alert threshold
		}),
	})

	// 2. CONFIGURE THE MAD ENGINE
	// MAD (Median Absolute Deviation) is an incredibly robust detector.
	// Unlike standard averages (which get confused by massive spikes),
	// MAD relies on the "Median", making it practically immune to extreme outliers.
	engineMAD := pipeline.NewEngine(pipeline.Config{
		WindowSize: 50,
		Detector: anomaly.NewMAD(anomaly.MADConfig{
			Threshold: 3.5, // Alert threshold
		}),
	})

	fmt.Println("--- Advanced Detectors (EWMA & MAD) ---")
	
	now := time.Now()
	for i := 0; i < 30; i++ {
		// Normal data fluctuating slightly so we have a non-zero variance
		val := 10.0 + float64(i%3)

		// Introduce a massive outlier on the 10th step
		if i == 10 {
			val = 1000.0
		} else if i == 25 {
			// Introduce a smaller anomaly on the 25th step
			val = 40.0
		}

		dp := timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Value:     val,
		}

		resEWMA := engineEWMA.Process(dp)
		resMAD := engineMAD.Process(dp)

		// Print only when one of them detects an anomaly
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
