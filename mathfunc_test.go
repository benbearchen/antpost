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
