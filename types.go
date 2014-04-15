package antpost

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type DroneResult int

const (
	ResultOK             DroneResult = iota
	ResultConnectFail    DroneResult = iota
	ResultResponseBroken DroneResult = iota
)

type DroneStep int

const (
	StepInit      DroneStep = iota
	StepConnected DroneStep = iota
	StepResponsed DroneStep = iota
)

type BoolReport struct {
	N            int
	True         int
	TruePercent  float32
	False        int
	FalsePercent float32
}

type DurationReport struct {
	N   int
	Avg time.Duration
	P05 time.Duration
	P50 time.Duration
	P95 time.Duration
}

type StatReport struct {
	Bools     map[string]*BoolReport
	Durations map[string]*DurationReport
	Subs      map[string]*StatReport
}

type Report struct {
	Time   *DurationReport
	OKTime *DurationReport
	Stat   *StatReport
}

func (r *Report) String() string {
	return "Time:    " + r.Time.String() + "\n" + "OKsTime: " + r.OKTime.String() + "\n" + "Stat >>>\n" + r.Stat.String()
}

func (s *StatReport) String() string {
	return "    " + strings.Join(s.string("    "), "\n    ") + "\n"
}

func (s *StatReport) string(prefix string) []string {
	r := make([]string, 0, len(s.Bools)+len(s.Durations))
	for n, b := range s.Bools {
		r = append(r, n+" \t"+b.String())
	}

	for n, d := range s.Durations {
		r = append(r, n+" \t"+d.String())
	}

	rs := make([]string, 0, len(s.Subs)*4)
	for n, s := range s.Subs {
		rs = append(rs, n+" >>>")
		for _, s := range s.string(prefix) {
			rs = append(rs, prefix+s)
		}
	}

	return append(r, rs...)
}

func (b *BoolReport) String() string {
	return fmt.Sprintf("n %7d,  true %7d(%.2f%%),  false %7d(%.2f%%)", b.N, b.True, b.TruePercent*100, b.False, b.FalsePercent*100)
}

func (d *DurationReport) String() string {
	return fmt.Sprintf("n %7d,  avg %v,  5%% %v,  50%% %v, 95%% %v", d.N, d.Avg, d.P05, d.P50, d.P95)
}

func AnalyzeBoolReport(values []bool) *BoolReport {
	n := len(values)
	if n <= 0 {
		return &BoolReport{N: 0}
	}

	t := 0
	for _, v := range values {
		if v {
			t++
		}
	}

	return &BoolReport{n, t, float32(t) / float32(n), n - t, float32(n-t) / float32(n)}
}

func AnalyzeDurationReport(times []time.Duration) *DurationReport {
	n := len(times)
	if n <= 0 {
		return &DurationReport{N: 0}
	}

	d := make([]int, len(times))
	for i, t := range times {
		d[i] = int(t / time.Microsecond)
	}

	sort.Sort(sort.IntSlice(d))

	var sum int64 = 0
	for _, d := range d {
		sum += int64(d)
	}

	p05 := time.Duration(d[n*5/100]) * time.Microsecond
	p50 := time.Duration(d[n*50/100]) * time.Microsecond
	p95 := time.Duration(d[n*95/100]) * time.Microsecond

	r := new(DurationReport)
	r.N = n
	r.Avg = time.Microsecond * time.Duration(sum/int64(n))
	r.P05 = p05
	r.P50 = p50
	r.P95 = p95
	return r
}
