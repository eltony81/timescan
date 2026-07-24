package main

import (
	"context"
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

	engines := make(map[string]*pipeline.Engine)
	for i := 1; i <= numServers; i++ {
		serverID := fmt.Sprintf("server-%02d", i)
		engines[serverID] = pipeline.NewEngine(pipeline.Config{
			WindowSize: 50,
			Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.5}),
		})
	}

	metricsChan := make(chan MetricPayload, 100)
	var wg sync.WaitGroup
	ctx := context.Background()

	go func() {
		for payload := range metricsChan {
			engine := engines[payload.ServerID]
			result, err := engine.Process(ctx, payload.Point)

			if err == nil && result.IsAnomaly {
				fmt.Printf("[ALERT] %s reported anomaly! Value: %.2f\n",
					payload.ServerID, payload.Point.Value)
			}
		}
	}()

	now := time.Now()
	for i := 1; i <= numServers; i++ {
		wg.Add(1)
		go func(serverNum int) {
			defer wg.Done()
			serverID := fmt.Sprintf("server-%02d", serverNum)

			for j := 0; j < numPointsPerServer; j++ {
				val := 50.0 + rand.Float64()*5.0

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
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	close(metricsChan)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("All metrics processed successfully across all servers.")
}
