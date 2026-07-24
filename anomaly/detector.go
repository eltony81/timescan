package anomaly

import "github.com/timescan/timescan/timeseries"

// AnomalyMeta holds additional information about a detected anomaly.
type AnomalyMeta struct {
	Score     float64
	Threshold float64
	Expected  float64
}

// AnomalyResult represents an anomaly found in a series.
type AnomalyResult struct {
	Point timeseries.DataPoint
	Meta  AnomalyMeta
}

// WindowContext provides contextual information about the time-series window.
type WindowContext struct {
	Window []timeseries.DataPoint
}

// Detector defines the unified interface for all anomaly detection algorithms.
type Detector interface {
	Detect(series timeseries.Series) []AnomalyResult
	IsAnomaly(point timeseries.DataPoint, ctx WindowContext) (bool, AnomalyMeta)
}
