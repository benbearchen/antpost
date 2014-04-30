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

type NominalReportItem struct {
	Name    string
	N       int
	Percent float32
}

type NominalReport struct {
	Items []*NominalReportItem
}

type OrdinalReportRank struct {
	Name                  string
	Order                 int
	N                     int
	Percent               float32
	CumulativeN           int
	CumulativePercent     float32
	DownCumulativeN       int
	DownCumulativePercent float32
}

type OrdinalReport struct {
	Ranks []*OrdinalReportRank
}

type IntervalReportItem struct {
	Value                 float64
	N                     int
	Percent               float32
	CumulativeN           int
	CumulativePercent     float32
	DownCumulativeN       int
	DownCumulativePercent float32
}

type IntervalReport struct {
	Interval float64
	Items    []*IntervalReportItem

	Mean              float64
	StandardDeviation float64 // 标准差
}

type RatioReport struct {
	N int

	Mean          float64 // 算术平均数
	GeometricMean float64 // 几何平均数
	QuadraticMean float64 // 平方平均数
	HarmonicMean  float64 // 调和平均数

	StandardDeviation float64 // 标准差

	P05 float64
	P25 float64
	P50 float64
	P75 float64
	P95 float64
}

type StatReport struct {
	Bools     map[string]*BoolReport
	Durations map[string]*DurationReport
	Subs      map[string]*StatReport
	Nominals  map[string]*NominalReport
	Ordinals  map[string]*OrdinalReport
	Intervals map[string]*IntervalReport
	Ratios    map[string]*RatioReport
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

	for n, v := range s.Nominals {
		rs = append(rs, n+" >>>")
		for _, s := range v.string() {
			rs = append(rs, prefix+s)
		}
	}

	for n, v := range s.Ordinals {
		rs = append(rs, n+" >>>")
		for _, s := range v.string() {
			rs = append(rs, prefix+s)
		}
	}

	for n, v := range s.Intervals {
		rs = append(rs, n+" >>>")
		for _, s := range v.string() {
			rs = append(rs, prefix+s)
		}
	}

	for n, v := range s.Ratios {
		rs = append(rs, n+" >>>")
		for _, s := range v.string() {
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

func (v *NominalReport) string() []string {
	r := make([]string, 0, len(v.Items))
	for _, item := range v.Items {
		s := fmt.Sprintf("%30s : n %7d(%6.2f%%)", item.Name, item.N, item.Percent)
		r = append(r, s)
	}

	return r
}

func (v *OrdinalReport) string() []string {
	r := make([]string, 0, len(v.Ranks)*2)
	for _, item := range v.Ranks {
		s := fmt.Sprintf("%10s : ord(%2d)", item.Name, item.Order)
		r = append(r, s)
		s = fmt.Sprintf("%12s n %6d(%6.2f%%), c %6d(%6.2f%%), d %6d(%6.2f%%)", "", item.N, item.Percent, item.CumulativeN, item.CumulativePercent, item.DownCumulativeN, item.DownCumulativePercent)
		r = append(r, s)
	}

	return r
}

func (v *IntervalReport) string() []string {
	r := make([]string, 0, len(v.Items)*2+1)
	s := fmt.Sprintf("interval: %15f, avg: %15f, sd: %15f", v.Interval, v.Mean, v.StandardDeviation)
	r = append(r, s)

	for _, item := range v.Items {
		s := fmt.Sprintf("step: %-15f(%d)", item.Value, int((item.Value-v.Items[0].Value)/v.Interval+0.01))
		r = append(r, s)
		s = fmt.Sprintf("   n %7d(%6.2f%%), c %7d(%6.2f%%), d %7d(%6.2f%%)", item.N, item.Percent, item.CumulativeN, item.CumulativePercent, item.DownCumulativeN, item.DownCumulativePercent)
		r = append(r, s)
	}

	return r
}

func (v *RatioReport) string() []string {
	r := make([]string, 0, 4)
	s := fmt.Sprintf("count: %13d,  avg: %15f,  SD: %16f", v.N, v.Mean, v.StandardDeviation)
	r = append(r, s)
	s = fmt.Sprintf("G: %17f,  Q: %17f,  H: %17f", v.GeometricMean, v.QuadraticMean, v.HarmonicMean)
	r = append(r, s)
	s = fmt.Sprintf("P05: %15f,  P25: %15f,  P25: %15f", v.P05, v.P25, v.P50)
	r = append(r, s)
	s = fmt.Sprintf("P75: %15f,  P95: %15f", v.P75, v.P95)
	r = append(r, s)

	return r
}
