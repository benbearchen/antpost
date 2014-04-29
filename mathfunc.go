package antpost

import (
	"math"
)

func calcMean(values []float64) float64 {
	var sum float64 = 0
	for _, v := range values {
		sum += v
	}

	if len(values) <= 0 {
		return sum
	} else {
		return sum / float64(len(values))
	}
}

func calcStandardDeviation(values []float64) float64 {
	if len(values) <= 0 {
		return 0
	}

	var sum float64 = 0
	var d float64 = 0
	for _, v := range values {
		sum += v
		d += v * v
	}

	avg := sum / float64(len(values))
	return math.Sqrt(d/float64(len(values)) - avg*avg)
}
