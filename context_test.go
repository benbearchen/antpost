package antpost

import "testing"

import (
	"fmt"
)

func TestStat(t *testing.T) {
	s := newStat()
	s.NominalInit("n", "A", "B", "C", "D")
	s.Nominal("n", "C")
	s.Nominal("n", "A")
	s.Nominal("n", "C")
	s.Nominal("n", "B")

	gen := NewOrdinalGen("0", "1", "2", "3")
	s.Ordinal("o", gen.Ord("0"))
	s.Ordinal("o", gen.Ord("1"))
	s.Ordinal("o", gen.Ord("1"))
	s.Ordinal("o", gen.Ord("3"))
	s.Ordinal("o", gen.Ord("3"))
	s.Ordinal("o", gen.Ord("3"))

	s.Interval("i", 1)
	s.Interval("i", 1.2)
	s.Interval("i", 1.5)
	s.Interval("i", 1.9)
	s.Interval("i", 1)

	s.IntervalInit("I", 0.25)
	s.Interval("I", 1.5)
	s.Interval("I", 2.5)
	s.Interval("I", 2.5)
	s.Interval("I", 4.0)

	s.Ratio("r", 5)
	s.Ratio("r", 6)
	s.Ratio("r", 8)
	s.Ratio("r", 9)

	r := s.Report()
	fmt.Println(r.String())

	if len(r.Nominals["n"].Items) != 4 {
		t.Errorf("Nominals failed")
	}

	if len(r.Ordinals["o"].Ranks) != 4 {
		t.Errorf("Ordinals failed")
	}

	if !isFloat64Approach(r.Intervals["i"].Interval, 0.1, 0.1) {
		t.Errorf("Intervals auto failed")
	}

	if !isFloat64Approach(r.Intervals["I"].Interval, 0.25, 0.25) {
		t.Errorf("Intervals set failed")
	}

	if !isFloat64Approach(r.Ratios["r"].Mean, 7, 7) {
		t.Errorf("Ratios failed")
	}
}
