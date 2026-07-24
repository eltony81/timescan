package anomaly

import (
	"math"

	"github.com/timescan/timescan/timeseries"
)

type MADConfig struct {
	Threshold float64 // typically 3.0 or 3.5
}

type MADDetector struct {
	config MADConfig
}

func NewMAD(config MADConfig) *MADDetector {
	return &MADDetector{config: config}
}

func (m *MADDetector) Detect(series timeseries.Series) []AnomalyResult {
	var results []AnomalyResult
	if len(series.Points) < 2 {
		return results
	}

	vals := make([]float64, len(series.Points))
	for i, p := range series.Points {
		vals[i] = p.Value
	}

	med := timeseries.Median(vals)

	devs := make([]float64, len(series.Points))
	for i, p := range series.Points {
		devs[i] = math.Abs(p.Value - med)
	}

	mad := timeseries.Median(devs)
	if mad == 0 {
		return results
	}

	for _, p := range series.Points {
		score := math.Abs(p.Value-med) / mad

		if score > m.config.Threshold {
			results = append(results, AnomalyResult{
				Point: p,
				Meta: AnomalyMeta{
					Score:     score,
					Threshold: m.config.Threshold,
					Expected:  med,
				},
			})
		}
	}
	return results
}

func (m *MADDetector) IsAnomaly(point timeseries.DataPoint, ctx WindowContext) (bool, AnomalyMeta) {
	if len(ctx.Window) < 2 {
		return false, AnomalyMeta{}
	}

	vals := make([]float64, len(ctx.Window))
	for i, p := range ctx.Window {
		vals[i] = p.Value
	}

	med := timeseries.Median(vals)
	devs := make([]float64, len(ctx.Window))
	for i, p := range ctx.Window {
		devs[i] = math.Abs(p.Value - med)
	}
	mad := timeseries.Median(devs)

	if mad == 0 {
		return false, AnomalyMeta{}
	}

	score := math.Abs(point.Value-med) / mad
	isAnomaly := score > m.config.Threshold

	return isAnomaly, AnomalyMeta{
		Score:     score,
		Threshold: m.config.Threshold,
		Expected:  med,
	}
}
