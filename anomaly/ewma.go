package anomaly

import (
	"math"

	"github.com/timescan/timescan/timeseries"
)

type EWMAConfig struct {
	Alpha     float64
	Threshold float64
}

type EWMADetector struct {
	config EWMAConfig
}

// NewEWMA creates a new EWMA detector. Clamps Alpha to (0, 1] if invalid.
func NewEWMA(config EWMAConfig) *EWMADetector {
	if config.Alpha <= 0 || config.Alpha > 1 {
		config.Alpha = 0.2
	}
	return &EWMADetector{config: config}
}

func (e *EWMADetector) Detect(series timeseries.Series) []AnomalyResult {
	var results []AnomalyResult
	if len(series.Points) < 2 {
		return results
	}

	ewma := series.Points[0].Value
	variance := 0.0

	for i, p := range series.Points {
		if i == 0 {
			continue
		}

		diff := p.Value - ewma
		stdDev := math.Sqrt(variance)

		if stdDev > 0 {
			score := math.Abs(diff) / stdDev
			if score > e.config.Threshold {
				results = append(results, AnomalyResult{
					Point: p,
					Meta: AnomalyMeta{
						Score:     score,
						Threshold: e.config.Threshold,
						Expected:  ewma,
					},
				})
			}
		}

		ewma = e.config.Alpha*p.Value + (1-e.config.Alpha)*ewma
		variance = e.config.Alpha*(diff*diff) + (1-e.config.Alpha)*variance
	}

	return results
}

func (e *EWMADetector) IsAnomaly(point timeseries.DataPoint, ctx WindowContext) (bool, AnomalyMeta) {
	if len(ctx.Window) < 1 {
		return false, AnomalyMeta{}
	}

	ewma := ctx.Window[0].Value
	variance := 0.0

	for i, p := range ctx.Window {
		if i == 0 {
			continue
		}
		diff := p.Value - ewma
		ewma = e.config.Alpha*p.Value + (1-e.config.Alpha)*ewma
		variance = e.config.Alpha*(diff*diff) + (1-e.config.Alpha)*variance
	}

	diff := point.Value - ewma
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return false, AnomalyMeta{}
	}

	score := math.Abs(diff) / stdDev
	isAnomaly := score > e.config.Threshold
	return isAnomaly, AnomalyMeta{
		Score:     score,
		Threshold: e.config.Threshold,
		Expected:  ewma,
	}
}
