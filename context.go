package antpost

import (
	"fmt"
	"time"
)

type Stat interface {
	Bool(name string, value bool)
	Duration(name string, duration time.Duration)
	Sub(name string) Stat
}

type Context struct {
	cur     *droneContext
	history []*droneContext
	count   int
	timer   *time.Timer
	stat    *stat
}

func NewContext() *Context {
	c := new(Context)
	c.history = make([]*droneContext, 0)
	c.count = -1
	c.stat = newStat()
	return c
}

func (c *Context) SetCount(count int) {
	c.count = count
	c.history = make([]*droneContext, 0, count+1)
}

func (c *Context) SetTime(d time.Duration) {
	c.timer = time.NewTimer(d)
}

func (c *Context) Combine(contexts ...*Context) {
	for _, v := range contexts {
		c.history = append(c.history, v.history...)
		c.stat.combine(v.stat)
	}
}

func (c *Context) Report() *Report {
	n := len(c.history)
	if n <= 0 {
		return nil
	}

	start := c.history[0].start
	end := c.history[0].end

	d := make([]time.Duration, 0, n)
	okd := make([]time.Duration, 0, n)
	for _, h := range c.history {
		if h.start.Before(start) {
			start = h.start
		}

		if end.After(h.end) {
			end = h.end
		}

		d = append(d, h.end.Sub(h.start))
		if h.step == StepResponsed && h.result == ResultOK {
			okd = append(okd, h.end.Sub(h.start))
		}
	}

	r := new(Report)
	r.Time = AnalyzeDurationReport(d)
	r.OKTime = AnalyzeDurationReport(okd)
	r.Stat = c.stat.Report()

	return r
}

func (c *Context) Start() bool {
	if c.count != 0 {
		if c.timer != nil {
			select {
			case <-c.timer.C:
				c.count = 0
				c.timer = nil
			default:
			}
		}
	}

	if c.count == 0 {
		return false
	} else if c.count > 0 {
		c.count--
	}

	c.cur = new(droneContext)
	c.cur.start = time.Now()
	return true
}

func (c *Context) Step(step DroneStep) {
	if c.cur == nil {
		panic(fmt.Errorf("Step() without Start()"))
	}

	switch c.cur.step {
	case StepInit:
		if step != StepConnected {
			panic(fmt.Errorf("Error step from StepInit"))
		}

		c.cur.step = step
		c.cur.connected = time.Now()
	case StepConnected:
		if step != StepResponsed {
			panic(fmt.Errorf("Error step from StepConnected"))
		}

		c.cur.step = step
		c.cur.responsed = time.Now()
	}
}

func (c *Context) End(result DroneResult) {
	if c.cur == nil {
		panic(fmt.Errorf("End() without Start()"))
	}

	c.cur.End(result)
	c.history = append(c.history, c.cur)
	c.cur = nil
}

func (c *Context) Bool(name string, value bool) {
	c.stat.Bool(name, value)
}

func (c *Context) Duration(name string, duration time.Duration) {
	c.stat.Duration(name, duration)
}

func (c *Context) SubStat(name string) Stat {
	return c.stat.Sub(name)
}

type droneContext struct {
	step      DroneStep
	start     time.Time
	connected time.Time
	responsed time.Time
	result    DroneResult
	end       time.Time
}

func (c *droneContext) End(result DroneResult) {
	c.result = result
	if !c.responsed.IsZero() {
		c.end = c.responsed
	} else if !c.connected.IsZero() {
		c.end = c.connected
	} else {
		c.end = time.Now()
	}
}

type stat struct {
	bools     map[string][]bool
	durations map[string][]time.Duration
	subs      map[string]*stat
}

func newStat() *stat {
	s := new(stat)
	s.bools = make(map[string][]bool)
	s.durations = make(map[string][]time.Duration)
	s.subs = make(map[string]*stat)
	return s
}

func (s *stat) Bool(name string, value bool) {
	v, ok := s.bools[name]
	if !ok {
		v = make([]bool, 0)
	}

	s.bools[name] = append(v, value)
}

func (s *stat) Duration(name string, duration time.Duration) {
	v, ok := s.durations[name]
	if !ok {
		v = make([]time.Duration, 0)
	}

	s.durations[name] = append(v, duration)
}

func (s *stat) Sub(name string) Stat {
	v, ok := s.subs[name]
	if !ok {
		v = newStat()
		s.subs[name] = v
	}

	return v
}

func (s *stat) Report() *StatReport {
	r := new(StatReport)
	r.Bools = make(map[string]*BoolReport)
	r.Durations = make(map[string]*DurationReport)
	r.Subs = make(map[string]*StatReport)

	for n, b := range s.bools {
		r.Bools[n] = AnalyzeBoolReport(b)
	}

	for n, d := range s.durations {
		r.Durations[n] = AnalyzeDurationReport(d)
	}

	for n, s := range s.subs {
		r.Subs[n] = s.Report()
	}

	return r
}

func (s *stat) combine(v *stat) {
	for n, b := range v.bools {
		a, ok := s.bools[n]
		if ok {
			a = append(a, b...)
		} else {
			a = b
		}

		s.bools[n] = a
	}

	for n, d := range v.durations {
		a, ok := s.durations[n]
		if ok {
			a = append(a, d...)
		} else {
			a = d
		}

		s.durations[n] = a
	}

	for n, s := range v.subs {
		a, ok := s.subs[n]
		if ok {
			a.combine(s)
		} else {
			s.subs[n] = a
		}
	}
}
