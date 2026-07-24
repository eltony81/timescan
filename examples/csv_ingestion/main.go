package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/timescan/timescan/anomaly"
	"github.com/timescan/timescan/timeseries"
)

func generateDummyCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"timestamp", "metric_value"})
	
	now := time.Now()
	for i := 0; i < 100; i++ {
		val := 50.0 + float64(i%5)
		if i == 75 {
			val = 200.0 // injected anomaly
		}
		ts := now.Add(time.Duration(i) * time.Minute).Format(time.RFC3339)
		writer.Write([]string{ts, fmt.Sprintf("%.2f", val)})
	}
	return nil
}

func main() {
	filename := "dataset.csv"
	generateDummyCSV(filename)
	defer os.Remove(filename) // clean up after test

	fmt.Println("--- CSV Ingestion & Backtesting ---")
	
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	_, _ = reader.Read()

	var points []timeseries.DataPoint
	for {
		record, err := reader.Read()
		if err != nil {
			break // EOF
		}
		
		ts, _ := time.Parse(time.RFC3339, record[0])
		val, _ := strconv.ParseFloat(record[1], 64)
		
		points = append(points, timeseries.DataPoint{
			Timestamp: ts,
			Value:     val,
		})
	}
	
	fmt.Printf("Successfully ingested %d rows from CSV.\n", len(points))

	// Run offline anomaly detection over the entire historical dataset
	series := timeseries.Series{Points: points}
	detector := anomaly.NewMAD(anomaly.MADConfig{Threshold: 3.5})
	
	anomalies := detector.Detect(series)
	fmt.Printf("Backtesting complete. Found %d historical anomalies.\n", len(anomalies))
	for _, a := range anomalies {
		fmt.Printf(" -> Anomaly at %s, Value: %.2f\n", a.Point.Timestamp.Format("15:04"), a.Point.Value)
	}
}
