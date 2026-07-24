package main

import (
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

func main() {
	// 1. CONFIGURE THE PIPELINE ENGINE
	// Think of the Engine as a conveyor belt for your incoming data.
	// We set up a "WindowSize" of 50. This means the engine will always remember
	// the most recent 50 data points (e.g., the last 50 seconds of CPU usage)
	// so it can understand what "normal" looks like.
	engine := pipeline.NewEngine(pipeline.Config{
		WindowSize: 50,
		
		// We use a Z-Score Detector. In simple terms, this calculates the average
		// of the recent data in the window. If a new number is drastically different 
		// from that average (specifically, 2 standard deviations away), it gets flagged!
		Detector: anomaly.NewZScore(anomaly.ZScoreConfig{
			Threshold: 2.0, 
		}),
		
		VectorStore: nil, // We don't need a database for this simple example.
	})

	fmt.Println("Starting Anomaly Detection Simulation...")
	fmt.Println("----------------------------------------")
	
	// 2. SIMULATE INCOMING DATA
	// Let's pretend we are receiving a new server reading (like temperature) every second.
	now := time.Now()
	for i := 0; i < 20; i++ {
		// Most of the time, the temperature is normal (around 100 or 105 degrees).
		val := 100.0
		if i == 15 {
			// Suddenly, at the 15th second, the temperature spikes dangerously to 250!
			val = 250.0 
		} else if i%2 == 0 {
			val = 105.0 // Just normal random fluctuations
		}

		// We package the reading into a "DataPoint" with the exact time and value.
		dp := timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Second),
			Value:     val,
			Tags:      map[string]string{"host": "server-01", "metric": "cpu_usage"},
		}

		// 3. PROCESS THE DATA
		// We feed the reading into our engine. The engine instantly compares it to the 
		// past history (the 50-point window) and tells us if this reading is dangerous.
		result := engine.Process(dp)

		timeStr := dp.Timestamp.Format("15:04:05")
		
		// 4. CHECK THE RESULT
		if result.IsAnomaly {
			// The engine successfully caught the sudden spike!
			fmt.Printf("[ALERT] 🚨 Anomaly Detected at %s!\n", timeStr)
			fmt.Printf("        Value: %.2f (The system expected around ~%.2f)\n", 
				dp.Value, result.AnomalyMeta.Expected)
		} else {
			// The data is normal, nothing to worry about.
			fmt.Printf("[OK]    Value: %.2f at %s\n", dp.Value, timeStr)
		}
	}
}
