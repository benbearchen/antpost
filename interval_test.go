package antpost

import "testing"

func doTestFraction(a, b float64, c, d int64, t *testing.T) {
	aa, bb := fraction(a, b)
	if aa != c || bb != d {
		t.Errorf("fraction(%v, %v) => %v, %v != %v, %v", a, b, aa, bb, c, d)
	}

	bb, aa = fraction(b, a)
	if aa != c || bb != d {
		t.Errorf("fraction(%v, %v) => %v, %v != %v, %v", b, a, bb, aa, d, c)
	}

}

func TestFraction(t *testing.T) {
	doTestFraction(0.1, 0.1, 1, 1, t)
	doTestFraction(1, 2, 1, 2, t)
	doTestFraction(0.1, 0.2, 1, 2, t)
	doTestFraction(0.3, 0.5, 3, 5, t)
	doTestFraction(0.03, 0.10, 3, 10, t)
	doTestFraction(0.13333, 0.4, 13333, 40000, t)
	doTestFraction(0.4/3, 0.4, 1, 3, t)
	doTestFraction(0.4/3, 0.13333333333333333333, 1, 1, t)
	doTestFraction(0.4e-14/3, 0.4e-14, 1, 3, t)
	doTestFraction(0.4e-14/3, 0.13333333333333333333e-14, 1, 1, t)
	doTestFraction(0.4e+14/3, 0.4e+14, 1, 3, t)
	doTestFraction(0.4e+14/3, 0.13333333333333333333e+14, 1, 1, t)
}

func doTestGCD(a, b, c int64, t *testing.T) {
	cc := gcd(a, b)
	if cc != c {
		t.Errorf("gcd(%v, %v) => %v != %v", a, b, cc, c)
	}

	cc = gcd(b, a)
	if cc != c {
		t.Errorf("gcd(%v, %v) => %v != %v", b, a, cc, c)
	}
}

func TestGCD(t *testing.T) {
	doTestGCD(1, 2, 1, t)
	doTestGCD(2, 2, 2, t)
	doTestGCD(3, 2, 1, t)
	doTestGCD(4, 2, 2, t)
	doTestGCD(5, 3, 1, t)
}

func doCalcIntervalInLCM(nums []float64, f float64, t *testing.T) {
	r := calcIntervalInLCM(nums[0], nums[1:])
	//if !isFloat64Approach(r, f, f) {
	if r != f {
		t.Errorf("calcIntervalInLCM(%v, %v) => %v != %v", nums[0], nums[1:], r, f)
	}
}

func TestCalcIntervalInLCM(t *testing.T) {
	doCalcIntervalInLCM([]float64{0.34}, 0.34, t)
	doCalcIntervalInLCM([]float64{0.34, 0.17}, 0.17, t)
	doCalcIntervalInLCM([]float64{0.07, 0.10}, 0.01, t)
	doCalcIntervalInLCM([]float64{0.07, 0.03}, 0.01, t)
	doCalcIntervalInLCM([]float64{1, 0.1, 0.2, 0.7}, 0.1, t)
	doCalcIntervalInLCM([]float64{100, 98, 17}, 1, t)
	m := 0.3425453
	doCalcIntervalInLCM([]float64{m, m / 7, m / 8, m / 9}, m/(7*8*9), t)
}

func doCalcInterval(values []float64, interval, min float64, t *testing.T) {
	i, m := calcInterval(values)
	//if i != interval || m != min {
	if !isFloat64Approach(i, interval, interval) || !isFloat64Approach(m, min, min) {
		t.Errorf("calcInterval(%v) => %v, %v != %v, %v", values, i, m, interval, min)
	}
}

func TestCalcInterval(t *testing.T) {
	doCalcInterval([]float64{1, 2, 3, 4, 5}, 1, 1, t)
	doCalcInterval([]float64{0.1, 0.2, 0.3, 0.4, 0.5}, 0.1, 0.1, t)
	doCalcInterval([]float64{0.13, 0.2, 0.3, 0.4, 0.5}, 0.01, 0.13, t)
	doCalcInterval([]float64{0.15, 0.2, 0.3, 0.4, 0.5}, 0.05, 0.15, t)
	m := 0.3425453
	doCalcInterval([]float64{m / 7, m / 8, m / 9}, m/(7*8*9), m/9, t)
	doCalcInterval([]float64{-m / 7, -m / 8, -m / 9}, m/(7*8*9), -m/7, t)
}

func doStepInterval(v, min, interval, r float64, t *testing.T) {
	result := stepInterval(v, min, interval)
	if result != r {
		t.Errorf("stepInterval(%v, %v, %v) => %v != %v", v, min, interval, result, r)
	}
}

func TestStepInterval(t *testing.T) {
	doStepInterval(1, 1, 1, 1, t)
	doStepInterval(2, 1, 1, 2, t)
	m := 0.3425453
	doStepInterval(m*3*0.999999+10, 10, m, m*3+10, t)
	doStepInterval(m*3*1.000001+10, 10, m, m*3+10, t)
}
