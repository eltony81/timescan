package timeseries

import "time"

// DataPoint represents a single metric observation in time.
type DataPoint struct {
	Timestamp time.Time
	Value     float64
	Tags      map[string]string
}

// Series represents a collection of data points.
type Series struct {
	Points []DataPoint
}
