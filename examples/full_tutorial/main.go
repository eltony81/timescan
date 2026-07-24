package main

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/decomposition"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
	"github.com/timescan/timescan/vector"
	"github.com/timescan/timescan/vector/driver/qdrant"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("   TIMESCAN MASTER TUTORIAL - FULL FEATURE DEMO   ")
	fmt.Println("==================================================")

	// ----------------------------------------------------
	// STEP 1: Setting up Vector Database Storage Adapter
	// ----------------------------------------------------
	fmt.Println("\n[Step 1] Initializing Vector DB Driver (Qdrant)...")
	vecStore, err := qdrant.NewStore(qdrant.Config{
		Addr:       "localhost:6334",
		Collection: "incident_patterns",
	})
	if err != nil {
		panic(err)
	}

	// ----------------------------------------------------
	// STEP 2: Creating Anomaly Detectors & Pipeline Engine
	// ----------------------------------------------------
	fmt.Println("[Step 2] Configuring Pipeline Engine with EWMA Detector...")
	detector := anomaly.NewEWMA(anomaly.EWMAConfig{
		Alpha:     0.15,
		Threshold: 2.5,
	})

	engine := pipeline.NewEngine(pipeline.Config{
		WindowSize:  60, // 60-point rolling window
		Detector:    detector,
		VectorStore: vecStore,
	})

	// ----------------------------------------------------
	// STEP 3: Streaming Ingestion & Real-Time Detection
	// ----------------------------------------------------
	fmt.Println("[Step 3] Simulating real-time metric ingestion with an anomaly...")
	now := time.Now()
	var historicalSeries timeseries.Series

	for i := 0; i < 60; i++ {
		// Generate base signal with slight noise + 7-period seasonality
		val := 100.0 + math.Sin(float64(i)*0.5)*15.0

		// Inject an anomaly spike at step 45
		if i == 45 {
			val = 350.0 // Sudden spike!
		}

		dp := timeseries.DataPoint{
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Value:     val,
			Tags:      map[string]string{"host": "prod-db-01", "region": "us-east-1"},
		}

		historicalSeries.Points = append(historicalSeries.Points, dp)

		// Evaluate in the hot-path (Zero-allocation internally)
		result := engine.Process(dp)

		if result.IsAnomaly {
			fmt.Printf("\n🚨 [ALERT AT STEP %d] Anomaly Detected!\n", i)
			fmt.Printf("   ├─ Timestamp:      %s\n", dp.Timestamp.Format("15:04:05"))
			fmt.Printf("   ├─ Observed Value: %.2f\n", dp.Value)
			fmt.Printf("   ├─ Expected Value: %.2f\n", result.AnomalyMeta.Expected)
			fmt.Printf("   └─ Anomaly Score:  %.2f (Threshold: %.2f)\n", 
				result.AnomalyMeta.Score, result.AnomalyMeta.Threshold)

			// ----------------------------------------------------
			// STEP 4: Vector Embedding & Similarity Pattern Match
			// ----------------------------------------------------
			fmt.Println("\n   [Step 4] Encoding Window Shape into Vector (PAA)...")
			dimensions := 8
			vectorEmbedding := vector.PAAEncode(result.WindowContext, dimensions)
			fmt.Printf("   ├─ Compressed 60-point window to %d dims: %v\n", dimensions, vectorEmbedding)

			// Upsert pattern to vector store
			incidentID := fmt.Sprintf("incident-%d", dp.Timestamp.Unix())
			_ = vecStore.Upsert(context.Background(), incidentID, vectorEmbedding, map[string]any{
				"host": dp.Tags["host"],
			})

			// Search for historical matches
			matches, _ := vecStore.SearchNearest(context.Background(), vectorEmbedding, 3, nil)
			fmt.Printf("   └─ Found %d similar past incidents in Qdrant.\n", len(matches))
		}
	}

	// ----------------------------------------------------
	// STEP 5: Offline Time-Series Seasonality Decomposition
	// ----------------------------------------------------
	fmt.Println("\n[Step 5] Running Offline Additive Decomposition on full series...")
	period := 7
	decomp := decomposition.DecomposeAdditive(historicalSeries, period)

	p45Raw := historicalSeries.Points[45].Value
	p45Trend := decomp.Trend.Points[45].Value
	p45Seasonal := decomp.Seasonal.Points[45].Value
	p45Residual := decomp.Residual.Points[45].Value

	fmt.Println("   Breakdown of Spike Point (Step 45):")
	fmt.Printf("   ├─ Raw Value:        %.2f\n", p45Raw)
	fmt.Printf("   ├─ Calculated Trend: %.2f\n", p45Trend)
	fmt.Printf("   ├─ Seasonal Effect:  %.2f\n", p45Seasonal)
	fmt.Printf("   └─ Residual Noise:   %.2f\n", p45Residual)

	// ----------------------------------------------------
	// STEP 6: Online Stats & Memory Recycling Summary
	// ----------------------------------------------------
	fmt.Println("\n[Step 6] Online Stats Summary (Welford & MAD)...")
	welford := timeseries.NewWelford()
	vals := make([]float64, len(historicalSeries.Points))
	for i, pt := range historicalSeries.Points {
		welford.Update(pt.Value)
		vals[i] = pt.Value
	}

	fmt.Printf("   ├─ Total Points Processed: %d\n", welford.Count())
	fmt.Printf("   ├─ Running Mean:           %.2f\n", welford.Mean())
	fmt.Printf("   ├─ Running StdDev:         %.2f\n", welford.StdDev())
	fmt.Printf("   ├─ Offline Median:         %.2f\n", timeseries.Median(vals))
	fmt.Printf("   └─ Offline MAD:            %.2f\n", timeseries.MAD(vals))

	fmt.Println("\n==================================================")
	fmt.Println("     TUTORIAL COMPLETED SUCCESSFULLY! 🎉          ")
	fmt.Println("==================================================")
}
