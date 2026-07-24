package main

import (
	"fmt"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

// AlertDispatcher handles debouncing of alerts to avoid spamming webhooks.
type AlertDispatcher struct {
	lastAlertTime time.Time
	cooldown      time.Duration
}

func (a *AlertDispatcher) Dispatch(dp timeseries.DataPoint, meta anomaly.AnomalyMeta) {
	if time.Since(a.lastAlertTime) < a.cooldown {
		fmt.Printf("[Throttled] Anomaly suppressed to prevent spam (Value: %.2f)\n", dp.Value)
		return
	}

	a.lastAlertTime = time.Now()
	// Simulate sending an HTTP POST request to Slack/Discord/Alertmanager
	fmt.Printf("[WEBHOOK] 🌐 POST /alerts -> 'CRITICAL: CPU Spike! Value: %.2f (Score: %.2f)'\n", 
		dp.Value, meta.Score)
}

func main() {
	engine := pipeline.NewEngine(pipeline.Config{
		WindowSize: 10,
		Detector:   anomaly.NewZScore(anomaly.ZScoreConfig{Threshold: 2.0}),
	})

	dispatcher := &AlertDispatcher{
		cooldown: 5 * time.Second, // Only send 1 alert every 5 seconds maximum
	}

	fmt.Println("--- Alert Webhook Dispatcher with Debouncing ---")
	
	// Simulate a prolonged incident (5 consecutive anomalous readings)
	for i := 0; i < 15; i++ {
		val := 10.0
		if i >= 5 && i <= 10 {
			val = 100.0 // Prolonged spike!
		} else {
			val = 10.0 + float64(i%2) // normal noise
		}

		dp := timeseries.DataPoint{
			Timestamp: time.Now(),
			Value:     val,
		}

		res := engine.Process(dp)
		if res.IsAnomaly {
			dispatcher.Dispatch(dp, res.AnomalyMeta)
		} else {
			fmt.Printf("[OK] Value: %.2f\n", val)
		}
		
		time.Sleep(1 * time.Second) // 1 second real-time delay to test debouncer
	}
}
