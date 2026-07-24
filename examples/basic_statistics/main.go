package main

import (
	"fmt"

	"github.com/timescan/timescan/timeseries"
)

func main() {
	// 1. ONLINE STATISTICS (Welford's Algorithm)
	// Welford's algorithm calculates Mean (Average), Variance, and Standard Deviation
	// "on the fly" as data arrives, without needing to store all the numbers in memory.
	// This is perfect for infinite streams of data!
	welford := timeseries.NewWelford()

	fmt.Println("--- Online Statistics (Welford) ---")
	
	streamData := []float64{10, 12, 23, 23, 16, 23, 21, 16}
	for _, val := range streamData {
		welford.Update(val) // Update stats instantly with minimal memory
		fmt.Printf("Added %.0f -> Current Mean: %.2f, StdDev: %.2f\n", 
			val, welford.Mean(), welford.StdDev())
	}
	
	fmt.Printf("Total items processed: %d\n\n", welford.Count())

	// 2. OFFLINE ROBUST STATISTICS
	// Sometimes you just have an array of data and want to find robust statistics
	// like the Median or the Interquartile Range (IQR). These metrics ignore
	// extreme outliers (unlike regular averages).
	data := []float64{1, 2, 3, 4, 100} // 100 is a massive outlier

	median := timeseries.Median(data) // The middle value
	iqr := timeseries.IQR(data)       // The spread of the middle 50%
	mad := timeseries.MAD(data)       // The median absolute deviation

	fmt.Println("--- Robust Array Statistics ---")
	fmt.Printf("Data Array: %v\n", data)
	fmt.Printf("Median:     %.2f (Notice how the 100 didn't ruin the center point!)\n", median)
	fmt.Printf("IQR:        %.2f\n", iqr)
	fmt.Printf("MAD:        %.2f\n", mad)
}
