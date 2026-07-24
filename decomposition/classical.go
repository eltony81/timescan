package decomposition

import (
	"github.com/timescan/timescan/timeseries"
)

// AdditiveDecomposition decomposes a series into Trend, Seasonal, and Residual components.
type AdditiveDecomposition struct {
	Trend    timeseries.Series
	Seasonal timeseries.Series
	Residual timeseries.Series
}

// DecomposeAdditive performs additive decomposition into Trend, Seasonal, and Residual.
// Returns empty decomposition if period <= 0 or series is too short.
func DecomposeAdditive(series timeseries.Series, period int) AdditiveDecomposition {
	if period <= 0 {
		return AdditiveDecomposition{}
	}

	n := len(series.Points)
	if n < period*2 {
		return AdditiveDecomposition{}
	}

	trend := computeMovingAverage(series, period)

	seasonalScores := make([]float64, period)
	counts := make([]float64, period)

	for i := 0; i < n; i++ {
		if trend.Points[i].Value != 0 {
			detrended := series.Points[i].Value - trend.Points[i].Value
			idx := i % period
			seasonalScores[idx] += detrended
			counts[idx]++
		}
	}

	for i := 0; i < period; i++ {
		if counts[i] > 0 {
			seasonalScores[i] /= counts[i]
		}
	}

	seasonal := timeseries.Series{Points: make([]timeseries.DataPoint, n)}
	residual := timeseries.Series{Points: make([]timeseries.DataPoint, n)}

	for i := 0; i < n; i++ {
		seasonal.Points[i] = timeseries.DataPoint{
			Timestamp: series.Points[i].Timestamp,
			Value:     seasonalScores[i%period],
			Tags:      series.Points[i].Tags,
		}

		var res float64
		if trend.Points[i].Value != 0 {
			res = series.Points[i].Value - trend.Points[i].Value - seasonal.Points[i].Value
		}

		residual.Points[i] = timeseries.DataPoint{
			Timestamp: series.Points[i].Timestamp,
			Value:     res,
			Tags:      series.Points[i].Tags,
		}
	}

	return AdditiveDecomposition{
		Trend:    trend,
		Seasonal: seasonal,
		Residual: residual,
	}
}

func computeMovingAverage(series timeseries.Series, period int) timeseries.Series {
	n := len(series.Points)
	res := timeseries.Series{Points: make([]timeseries.DataPoint, n)}

	half := period / 2
	for i := half; i < n-half; i++ {
		sum := 0.0
		for j := i - half; j <= i+half; j++ {
			sum += series.Points[j].Value
		}
		res.Points[i] = timeseries.DataPoint{
			Timestamp: series.Points[i].Timestamp,
			Value:     sum / float64(period),
			Tags:      series.Points[i].Tags,
		}
	}
	return res
}
