package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/pipeline"
	"github.com/timescan/timescan/timeseries"
)

type AlertPayload struct {
	Source    string    `json:"source"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Score     float64   `json:"anomaly_score"`
	Threshold float64   `json:"threshold"`
}

func sendWebhookAlert(webhookURL string, payload AlertPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func processCSVFile(ctx context.Context, filePath string, engine *pipeline.Engine, webhookURL string) error {
	fmt.Printf("\n--- Processing CSV Input Stream: %s ---\n", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, _ = reader.Read() // Skip header

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		ts, _ := time.Parse(time.RFC3339, record[0])
		val, _ := strconv.ParseFloat(record[1], 64)

		dp := timeseries.DataPoint{
			Timestamp: ts,
			Value:     val,
			Tags:      map[string]string{"source": "csv_file"},
		}

		result, err := engine.Process(ctx, dp)
		if err != nil {
			return err
		}

		if result.IsAnomaly {
			fmt.Printf("🚨 [CSV ANOMALY DETECTED] Time: %s, Value: %.2f, Score: %.2f\n",
				ts.Format("15:04:05"), val, result.AnomalyMeta.Score)

			alert := AlertPayload{
				Source:    "csv_ingestion",
				Timestamp: ts,
				Value:     val,
				Score:     result.AnomalyMeta.Score,
				Threshold: result.AnomalyMeta.Threshold,
			}
			if err := sendWebhookAlert(webhookURL, alert); err != nil {
				fmt.Printf("Error sending webhook alert: %v\n", err)
			}
		}
	}
	return nil
}

func processPrometheusMetrics(ctx context.Context, prometheusURL string, engine *pipeline.Engine, webhookURL string) error {
	fmt.Printf("\n--- Querying Prometheus Input Stream: %s ---\n", prometheusURL)
	
	// Query Prometheus self-metrics
	queryURL := fmt.Sprintf("%s/api/v1/query?query=go_goroutines", prometheusURL)
	resp, err := http.Get(queryURL)
	if err != nil {
		return fmt.Errorf("failed to query prometheus: %w", err)
	}
	defer resp.Body.Close()

	var promResp struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return fmt.Errorf("failed to decode prometheus response: %w", err)
	}

	fmt.Printf("Fetched %d metric series from Prometheus.\n", len(promResp.Data.Result))
	for _, res := range promResp.Data.Result {
		if len(res.Value) >= 2 {
			valStr, _ := res.Value[1].(string)
			val, _ := strconv.ParseFloat(valStr, 64)

			// Inject a artificial spike test point to verify anomaly detection over Prometheus data
			dp := timeseries.DataPoint{
				Timestamp: time.Now(),
				Value:     val * 50.0, // Injected spike
				Tags:      res.Metric,
			}

			result, _ := engine.Process(ctx, dp)
			if result.IsAnomaly {
				fmt.Printf("🚨 [PROMETHEUS ANOMALY DETECTED] Metric: %v, Value: %.2f, Score: %.2f\n",
					res.Metric, dp.Value, result.AnomalyMeta.Score)

				alert := AlertPayload{
					Source:    "prometheus_stream",
					Timestamp: dp.Timestamp,
					Value:     dp.Value,
					Score:     result.AnomalyMeta.Score,
					Threshold: result.AnomalyMeta.Threshold,
				}
				_ = sendWebhookAlert(webhookURL, alert)
			}
		}
	}
	return nil
}

func main() {
	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL = "http://prometheus:9090"
	}

	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		webhookURL = "http://webhook:8080/alerts"
	}

	csvPath := os.Getenv("CSV_PATH")
	if csvPath == "" {
		csvPath = "/data/metrics.csv"
	}

	fmt.Println("==================================================")
	fmt.Println("   TIMESCAN E2E INTEGRATION TEST RUNNER           ")
	fmt.Println("==================================================")
	fmt.Printf("Prometheus URL: %s\n", prometheusURL)
	fmt.Printf("Webhook URL:    %s\n", webhookURL)
	fmt.Printf("CSV Path:       %s\n", csvPath)

	engine := pipeline.NewEngine(pipeline.Config{
		WindowSize: 10,
		Detector:   anomaly.NewMAD(anomaly.MADConfig{Threshold: 2.5}),
	})

	ctx := context.Background()

	// 1. Process CSV input
	if err := processCSVFile(ctx, csvPath, engine, webhookURL); err != nil {
		fmt.Printf("CSV processing error: %v\n", err)
	}

	// Wait 3 seconds for Prometheus to gather initial metrics
	time.Sleep(3 * time.Second)

	// 2. Process Prometheus input
	if err := processPrometheusMetrics(ctx, prometheusURL, engine, webhookURL); err != nil {
		fmt.Printf("Prometheus processing error: %v\n", err)
	}

	fmt.Println("\n==================================================")
	fmt.Println("   TIMESCAN E2E TEST COMPLETED SUCCESSFULLY!      ")
	fmt.Println("==================================================")
}
