package antpost

import (
	"math"
	"sort"
)

func calcInterval(values []float64) (interval, min float64) {
	s := uniqueFloat64s(values)
	sort.Float64s(s)

	min = s[0]
	abs := math.Max(math.Abs(s[0]), math.Abs(s[len(s)-1]))

	d := make([]float64, 0, len(s)-1)
	for i := 1; i < len(s); i++ {
		if isFloat64Approach(s[i], s[i-1], abs) {
			continue
		}

		d = append(d, s[i]-s[i-1])
	}

	d = uniqueFloat64s(d)
	sort.Float64s(d)
	if len(d) == 0 {
		return 1, min
	} else if len(d) == 1 {
		return d[0], min
	}

	r := make([]float64, 0, len(d)-1)
	for i := 0; i < cap(r); i++ {
		f := math.Mod(d[i+1], d[0])
		if m := f / d[0]; m > 1e-14 && m < 1-1e-14 {
			r = append(r, f)
		}
	}

	r = uniqueFloat64s(r)

	interval = calcIntervalInLCM(d[0], r)
	return interval, min
}

func fraction(dem, num float64) (denominator, numerator int64) {
	d, f := dem, num
	if d < f {
		d, f = f, d
	}

	for d > f {
		m := math.Mod(d, f)
		if isFloat64ApproachZero(m, dem) {
			break
		}

		d, f = f, m
	}

	return divToInt(dem, f), divToInt(num, f)
}

func divToInt(d, f float64) int64 {
	return int64(math.Floor(d/f + 1e-14))
}

func gcd(d, f int64) int64 {
	if d < f {
		d, f = f, d
	}

	for d > f {
		m := d % f
		if m == 0 {
			break
		}

		d, f = f, m
	}

	return f
}

func calcIntervalInLCM(min float64, d []float64) float64 {
	var lcm int64 = 1
	for _, d := range d {
		dem, _ := fraction(min, d)
		lcm = lcm / gcd(lcm, dem) * dem
	}

	return min / float64(lcm)
}

func isFloat64Approach(a, b, max float64) bool {
	return math.Abs(a-b) <= math.Abs(max)*1e-14
}

func isFloat64ApproachZero(v, max float64) bool {
	return math.Abs(v) <= math.Abs(max)*1e-7
}

func uniqueFloat64s(values []float64) []float64 {
	f := make(map[float64]bool)
	for _, v := range values {
		if _, ok := f[v]; !ok {
			f[v] = true
		}
	}

	s := make([]float64, 0, len(f))
	for v, _ := range f {
		s = append(s, v)
	}

	return s
}

func stepInterval(v, min, interval float64) float64 {
	return min + interval*math.Floor((v-min)/interval+0.1)
}
