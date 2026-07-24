package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

type MetricPayload struct {
	ServerID string
	Point    timeseries.DataPoint
}

func main() {
	fmt.Println("--- High-Performance Concurrent Monitoring ---")
	
	numServers := 5
	numPointsPerServer := 20
	
	// Create a worker pool pipeline for each server
	engines := make(map[string]*pipeline.Engine)
	for i := 1; i <= numServers; i++ {
		serverID := fmt.Sprintf("server-%02d", i)
		engines[serverID] = pipeline.NewEngine(pipeline.Config{
			WindowSize: 50,
			Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.5}),
		})
	}

	// Channel to ingest metrics from all servers concurrently
	metricsChan := make(chan MetricPayload, 100)
	var wg sync.WaitGroup

	// Start a central dispatcher to process incoming metrics
	go func() {
		for payload := range metricsChan {
			// Route payload to the corresponding engine
			engine := engines[payload.ServerID]
			result := engine.Process(payload.Point)
			
			if result.IsAnomaly {
				fmt.Printf("[ALERT] %s reported anomaly! Value: %.2f\n", 
					payload.ServerID, payload.Point.Value)
			}
		}
	}()

	// Simulate N servers sending data concurrently
	now := time.Now()
	for i := 1; i <= numServers; i++ {
		wg.Add(1)
		go func(serverNum int) {
			defer wg.Done()
			serverID := fmt.Sprintf("server-%02d", serverNum)
			
			for j := 0; j < numPointsPerServer; j++ {
				val := 50.0 + rand.Float64()*5.0
				
				// Server 3 randomly crashes/spikes on step 10
				if serverNum == 3 && j == 10 {
					val = 99.0
				}

				metricsChan <- MetricPayload{
					ServerID: serverID,
					Point: timeseries.DataPoint{
						Timestamp: now.Add(time.Duration(j) * time.Second),
						Value:     val,
					},
				}
				time.Sleep(10 * time.Millisecond) // simulate network delay
			}
		}(i)
	}

	wg.Wait()
	close(metricsChan)
	// Allow dispatcher to flush
	time.Sleep(100 * time.Millisecond)
	fmt.Println("All metrics processed successfully across all servers.")
}
