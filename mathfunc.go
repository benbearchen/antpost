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
		return 0
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

func calcGeometricMean(values []float64) float64 {
	var product float64 = 1
	c := 0
	for _, v := range values {
		if v > 0 {
			product *= v
			c++
		}
	}

	if c > 0 {
		return math.Pow(product, 1.0/float64(c))
	} else {
		return 0
	}
}

func calcQuadraticMean(values []float64) float64 {
	if len(values) <= 0 {
		return 0
	}

	var sum float64 = 0
	for _, v := range values {
		sum += v * v
	}

	return math.Sqrt(sum / float64(len(values)))
}

func calcHarmonicMean(values []float64) float64 {
	var sum float64 = 0
	c := 0
	for _, v := range values {
		if v > 0 {
			sum += 1 / v
			c++
		}
	}

	if c > 0 {
		return float64(c) / sum
	} else {
		return 0
	}
}
