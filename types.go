package antpost

import (
	"sort"
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

type DurationReport struct {
	N   int
	Avg time.Duration
	P05 time.Duration
	P50 time.Duration
	P95 time.Duration
}

type Report struct {
	Time   DurationReport
	OKTime DurationReport
}

func (r *DurationReport) Analyze(times []time.Duration) {
	n := len(times)
	if n <= 0 {
		return
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

	r.N = n
	r.Avg = time.Microsecond * time.Duration(sum/int64(n))
	r.P05 = p05
	r.P50 = p50
	r.P95 = p95
}
