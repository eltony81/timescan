package timeseries

import (
	"math"
	"testing"
)

func TestStats(t *testing.T) {
	t.Run("Median", func(t *testing.T) {
		data := []float64{3, 1, 2}
		med := Median(data)
		if med != 2 {
			t.Errorf("expected 2, got %f", med)
		}
		if data[0] != 3 {
			t.Errorf("input slice was mutated: %v", data)
		}
	})

	t.Run("IQR", func(t *testing.T) {
		data := []float64{1, 2, 5, 6, 7, 9, 12, 15}
		iqr := IQR(data)
		if iqr <= 0 {
			t.Errorf("expected positive IQR, got %f", iqr)
		}
	})

	t.Run("MAD", func(t *testing.T) {
		data := []float64{1, 1, 2, 2, 4, 6, 9}
		mad := MAD(data)
		if mad <= 0 {
			t.Errorf("expected positive MAD, got %f", mad)
		}
	})
}

func FuzzWelford(f *testing.F) {
	f.Add(1.0, 2.0, 3.0)
	f.Fuzz(func(t *testing.T, a, b, c float64) {
		w := NewWelford()
		w.Update(a)
		w.Update(b)
		w.Update(c)
		if w.Count() > 0 && math.IsNaN(w.Mean()) && !math.IsNaN(a) && !math.IsNaN(b) && !math.IsNaN(c) {
			t.Errorf("unexpected NaN mean")
		}
	})
}

func FuzzMedian(f *testing.F) {
	f.Add(1.0, 2.0, 3.0, 4.0)
	f.Fuzz(func(t *testing.T, a, b, c, d float64) {
		slice := []float64{a, b, c, d}
		_ = Median(slice)
		_ = IQR(slice)
		_ = MAD(slice)
	})
}
