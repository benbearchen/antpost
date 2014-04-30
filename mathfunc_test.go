package antpost

import "testing"

import (
	"math"
)

func TestMean(t *testing.T) {
	values := []float64{5, 6, 8, 9}
	mean := 7.0
	if m := calcMean(values); m != mean {
		t.Errorf("calcMean(%v) => %v != %v", values, m, mean)
	}
}

func TestStandardDeviation(t *testing.T) {
	values := []float64{5, 6, 8, 9}
	sd := math.Sqrt(2.5)
	if d := calcStandardDeviation(values); d != sd {
		t.Errorf("calcStandardDeviation(%v) => %v != %v", values, d, sd)
	}
}

func TestGeometricMean(t *testing.T) {
	values := []float64{4, 1, 1.0 / 32}
	gm := 1.0 / 2
	if g := calcGeometricMean(values); !isFloat64Approach(g, gm, gm) {
		t.Errorf("calcGeometricMean(%v) => %v != %v", values, g, gm)
	}
}

func TestQuadraticMean(t *testing.T) {
	values := []float64{-2, -1, 1, 2}
	qm := math.Sqrt(2.5)
	if q := calcQuadraticMean(values); q != qm {
		t.Errorf("calcQuadraticMean(%v) => %v != %v", values, q, qm)
	}
}

func TestHarmonicMean(t *testing.T) {
	values := []float64{1, 2, 4}
	hm := 12.0 / 7
	if h := calcHarmonicMean(values); h != hm {
		t.Errorf("calcHarmonicMean(%v) => %v != %v", values, h, hm)
	}
}
