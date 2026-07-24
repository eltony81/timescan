package main

import (
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/decomposition"
	"github.com/timescan/timescan/timeseries"
)

func main() {
	// We have a signal with a strong weekly pattern (period = 7)
	// We want to detect an anomaly that is hidden inside a normal spike.
	n := 35
	period := 7
	series := timeseries.Series{Points: make([]timeseries.DataPoint, n)}
	now := time.Now()

	for i := 0; i < n; i++ {
		val := 100.0
		// Weekly spike of +50 on weekends
		if i%period == 5 || i%period == 6 {
			val += 50.0
		}

		// Introduce a hidden anomaly on day 15 (a Tuesday).
		// Normally a Tuesday should be ~100. If it's 140, it's weird!
		// But 140 is lower than a normal weekend (150). So a basic threshold might miss it.
		if i == 15 {
			val = 140.0
		}

		series.Points[i] = timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * 24 * time.Hour),
			Value:     val,
		}
	}

	// 1. Decompose the series to isolate the Residual noise
	decomp := decomposition.DecomposeAdditive(series, period)

	// 2. Run Anomaly Detection ONLY on the Residual
	detector := anomaly.NewMAD(anomaly.MADConfig{Threshold: 3.0})
	anomalies := detector.Detect(decomp.Residual)

	fmt.Println("--- Seasonal Anomaly Detection ---")
	fmt.Println("We injected an anomaly of 140 on Day 15 (Tuesday).")
	fmt.Println("A normal weekend reaches 150. A basic detector would ignore 140.")
	fmt.Printf("Anomalies found in Residual by MAD: %d\n", len(anomalies))

	for _, a := range anomalies {
		// Find the original day index for printing
		var dayIndex int
		var originalVal float64
		for i, p := range series.Points {
			if p.Timestamp.Equal(a.Point.Timestamp) {
				dayIndex = i
				originalVal = p.Value
			}
		}
		fmt.Printf("🚨 Anomaly triggered on Day %d! Original Value: %.2f (Residual score: %.2f)\n",
			dayIndex, originalVal, a.Meta.Score)
	}
}
