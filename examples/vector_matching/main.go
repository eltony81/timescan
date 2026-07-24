package main

import (
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/timeseries"
	"github.com/timescan/timescan/vector"
	"github.com/timescan/timescan/vector/driver/qdrant"
)

func main() {
	// 1. INITIALIZE A VECTOR DATABASE (QDRANT)
	// A vector database stores mathematical representations of shapes (like graph lines).
	// When our system crashes, we can search this database to find if an identical
	// graphical pattern happened in the past.
	vecStore, _ := qdrant.NewStore(qdrant.Config{
		Addr:       "localhost:6334",
		Collection: "patterns",
	})
	_ = vecStore // In a real app, you would use this to search or save the patterns.

	// 2. PREPARE THE HISTORICAL DATA
	// Imagine we just detected a massive anomaly. We grab the last 60 minutes
	// of data leading up to the crash so we can analyze the "shape" of the disaster.
	windowSize := 60
	points := make([]timeseries.DataPoint, windowSize)
	now := time.Now()

	for i := 0; i < windowSize; i++ {
		points[i] = timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Value:     float64(i%10) * 1.5, // We simulate a wavy, periodic graph line.
		}
	}

	// We bundle these 60 raw points into a "Context" window.
	ctx := anomaly.WindowContext{
		Window: points,
	}

	// 3. COMPRESS THE DATA (PAA ENCODING)
	// Databases cannot easily compare 60 raw data points.
	// Instead, we use an algorithm called PAA (Piecewise Aggregate Approximation)
	// to neatly compress these 60 points into exactly 8 summary numbers (dimensions).
	// Think of it like taking a high-resolution image and blurring it slightly
	// so the computer can analyze it much faster.
	dimensions := 8
	encodedVec := vector.PAAEncode(ctx, dimensions)

	fmt.Println("--- Vector Pattern Matching Example ---")
	fmt.Printf("Original Graph Data: %d individual data points\n", windowSize)
	fmt.Printf("Compressed Vector (%d dimensions): %v\n", dimensions, encodedVec)
	fmt.Println("\nSuccess! This small array of 8 numbers can now be sent to Qdrant.")
	fmt.Println("Qdrant will instantly search its memory to find if this exact graphical 'shape'")
	fmt.Println("has ever caused a system crash before!")
}
