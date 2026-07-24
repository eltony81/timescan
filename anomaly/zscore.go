package anomaly

import (
	"math"

	"github.com/timescan/timescan/timeseries"
)

type ZScoreConfig struct {
	Threshold float64
}

type ZScoreDetector struct {
	config ZScoreConfig
}

func NewZScore(config ZScoreConfig) *ZScoreDetector {
	return &ZScoreDetector{config: config}
}

func (z *ZScoreDetector) Detect(series timeseries.Series) []AnomalyResult {
	if len(series.Points) < 2 {
		return nil
	}

	w := timeseries.NewWelford()
	for _, p := range series.Points {
		w.Update(p.Value)
	}
	mean := w.Mean()
	stdDev := w.StdDev()

	var results []AnomalyResult
	if stdDev == 0 {
		return results
	}

	for _, p := range series.Points {
		score := math.Abs(p.Value-mean) / stdDev
		if score > z.config.Threshold {
			results = append(results, AnomalyResult{
				Point: p,
				Meta: AnomalyMeta{
					Score:     score,
					Threshold: z.config.Threshold,
					Expected:  mean,
				},
			})
		}
	}
	return results
}

func (z *ZScoreDetector) IsAnomaly(point timeseries.DataPoint, ctx WindowContext) (bool, AnomalyMeta) {
	if len(ctx.Window) < 2 {
		return false, AnomalyMeta{}
	}
	w := timeseries.NewWelford()
	for _, p := range ctx.Window {
		w.Update(p.Value)
	}
	mean := w.Mean()
	stdDev := w.StdDev()

	if stdDev == 0 {
		return false, AnomalyMeta{}
	}

	score := math.Abs(point.Value-mean) / stdDev
	isAnomaly := score > z.config.Threshold

	return isAnomaly, AnomalyMeta{
		Score:     score,
		Threshold: z.config.Threshold,
		Expected:  mean,
	}
}
