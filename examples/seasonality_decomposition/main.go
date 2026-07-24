package main

import (
	"fmt"
	"time"

	"github.com/timescan/timescan/decomposition"
	"github.com/timescan/timescan/timeseries"
)

func main() {
	// 1. CREATE A TIME SERIES WITH TREND AND SEASONALITY
	// Imagine tracking the daily sales of a coffee shop. 
	// - Trend: Sales are slowly going up over time as the shop gets popular.
	// - Seasonality: Every 7 days (a week), there is a predictable spike on weekends.
	// - Residual: Random noise (some rainy days, some lucky days).
	
	n := 30 // 30 days of data
	period := 7 // A repeating pattern every 7 days
	
	series := timeseries.Series{
		Points: make([]timeseries.DataPoint, n),
	}

	now := time.Now()
	for i := 0; i < n; i++ {
		// Base value
		val := 100.0
		// Trend: slowly growing by 2.0 every day
		trend := float64(i) * 2.0
		// Seasonality: spikes on day 5 and 6 (the weekend)
		seasonality := 0.0
		if i%period == 5 || i%period == 6 {
			seasonality = 50.0 // Weekend spike!
		}
		
		val = val + trend + seasonality

		series.Points[i] = timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * 24 * time.Hour),
			Value:     val,
		}
	}

	// 2. PERFORM ADDITIVE DECOMPOSITION
	// We ask the library to break down our raw graph into its three root components:
	// 1. The underlying Trend (is the business growing?)
	// 2. The Seasonality (what is the predictable repeating pattern?)
	// 3. The Residual (the random noise left over)
	result := decomposition.DecomposeAdditive(series, period)

	fmt.Println("--- Seasonality Decomposition Example ---")
	fmt.Printf("Analyzed %d days of data with a period of %d days.\n\n", n, period)

	// Let's print the middle point (day 15) to see how it was broken down
	middleIndex := 15
	raw := series.Points[middleIndex].Value
	trend := result.Trend.Points[middleIndex].Value
	season := result.Seasonal.Points[middleIndex].Value
	residual := result.Residual.Points[middleIndex].Value

	fmt.Printf("Day %d Analysis:\n", middleIndex)
	fmt.Printf("Raw Value:          %.2f\n", raw)
	fmt.Printf(" ├─ Underlying Trend: %.2f (General growth)\n", trend)
	fmt.Printf(" ├─ Seasonal Effect:  %.2f (The weekly pattern)\n", season)
	fmt.Printf(" └─ Residual Noise:   %.2f (Random unexplained variance)\n", residual)
	
	// Notice that raw ≈ trend + season + residual
	fmt.Printf("\nMath Check: %.2f + %.2f + %.2f = %.2f\n", trend, season, residual, trend+season+residual)
}
