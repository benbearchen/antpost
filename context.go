package antpost

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type OrdinalRank interface {
	Name() string
	Order() int
	All() []OrdinalRank
	Related(rank OrdinalRank) bool
}

type OrdinalGen interface {
	Ord(rank string) OrdinalRank
}

func NewOrdinalGen(ranks ...string) OrdinalGen {
	return newOrdinalGen(ranks...)
}

type Stat interface {
	Bool(name string, value bool)
	Duration(name string, duration time.Duration)
	Sub(name string) Stat

	// 4 kinds of `scale of measure', see  http://zh.wikipedia.org/wiki/%E6%B8%AC%E9%87%8F%E7%9A%84%E5%B0%BA%E5%BA%A6
	Nominal(name string, item string)
	NominalInit(name string, items ...string)
	Ordinal(name string, rank OrdinalRank)
	Interval(name string, value float64)
	IntervalInit(name string, interval float64)
	Ratio(name string, value float64)
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

func (c *Context) Stat() Stat {
	return c.stat
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
	nominals  map[string]*nominalStat
	ordinals  map[string]*ordinalStat
	intervals map[string]*intervalStat
	ratios    map[string]*ratioStat
}

func newStat() *stat {
	s := new(stat)
	s.bools = make(map[string][]bool)
	s.durations = make(map[string][]time.Duration)
	s.subs = make(map[string]*stat)
	s.nominals = make(map[string]*nominalStat)
	s.ordinals = make(map[string]*ordinalStat)
	s.intervals = make(map[string]*intervalStat)
	s.ratios = make(map[string]*ratioStat)
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

func (s *stat) Nominal(name string, item string) {
	n, ok := s.nominals[name]
	if !ok {
		n = newNominalStat()
		s.nominals[name] = n
	}

	n.Item(item)
}

func (s *stat) NominalInit(name string, items ...string) {
	n, ok := s.nominals[name]
	if !ok {
		n = newNominalStat()
		s.nominals[name] = n
	}

	n.Init(items)
}

func (s *stat) Ordinal(name string, rank OrdinalRank) {
	ord, ok := s.ordinals[name]
	if !ok {
		ord = newOrdinalStat(rank.All())
		s.ordinals[name] = ord
	}

	ord.Rank(rank)
}

func (s *stat) Interval(name string, value float64) {
	i, ok := s.intervals[name]
	if !ok {
		i = newIntervalStat()
		s.intervals[name] = i
	}

	i.Value(value)
}

func (s *stat) IntervalInit(name string, interval float64) {
	i, ok := s.intervals[name]
	if !ok {
		i = newIntervalStat()
		s.intervals[name] = i
	}

	i.Init(interval)
}

func (s *stat) Ratio(name string, value float64) {
	r, ok := s.ratios[name]
	if !ok {
		r = newRatioStat()
		s.ratios[name] = r
	}

	r.Value(value)
}

func (s *stat) Report() *StatReport {
	r := new(StatReport)
	r.Bools = make(map[string]*BoolReport)
	r.Durations = make(map[string]*DurationReport)
	r.Subs = make(map[string]*StatReport)
	r.Nominals = make(map[string]*NominalReport)
	r.Ordinals = make(map[string]*OrdinalReport)
	r.Intervals = make(map[string]*IntervalReport)
	r.Ratios = make(map[string]*RatioReport)

	for n, b := range s.bools {
		r.Bools[n] = AnalyzeBoolReport(b)
	}

	for n, d := range s.durations {
		r.Durations[n] = AnalyzeDurationReport(d)
	}

	for n, s := range s.subs {
		r.Subs[n] = s.Report()
	}

	for n, v := range s.nominals {
		r.Nominals[n] = v.report()
	}

	for n, v := range s.ordinals {
		r.Ordinals[n] = v.report()
	}

	for n, v := range s.intervals {
		r.Intervals[n] = v.report()
	}

	for n, v := range s.ratios {
		r.Ratios[n] = v.report()
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

	for n, sub := range v.subs {
		a, ok := s.subs[n]
		if ok {
			a.combine(sub)
		} else {
			s.subs[n] = sub
		}
	}

	for n, nominal := range v.nominals {
		a, ok := s.nominals[n]
		if ok {
			a.combine(nominal)
		} else {
			s.nominals[n] = nominal
		}
	}

	for n, ordinal := range v.ordinals {
		a, ok := s.ordinals[n]
		if ok {
			a.combine(ordinal)
		} else {
			s.ordinals[n] = ordinal
		}
	}

	for n, interval := range v.intervals {
		a, ok := s.intervals[n]
		if ok {
			a.combine(interval)
		} else {
			s.intervals[n] = interval
		}
	}

	for n, ratio := range v.ratios {
		a, ok := s.ratios[n]
		if ok {
			a.combine(ratio)
		} else {
			s.ratios[n] = ratio
		}
	}
}

type ordinalGen struct {
	ranks map[string]int
}

type ordinalRank struct {
	gen  *ordinalGen
	rank string
}

func newOrdinalGen(ranks ...string) OrdinalGen {
	gen := new(ordinalGen)
	gen.ranks = make(map[string]int)
	for i, r := range ranks {
		gen.ranks[r] = i
	}

	return gen
}

func (g *ordinalGen) Ord(rank string) OrdinalRank {
	if _, ok := g.ranks[rank]; ok {
		return &ordinalRank{g, rank}
	} else {
		panic(fmt.Errorf("ordinalGen(%v) has not %s", g.ranks, rank))
	}
}

func (r *ordinalRank) Name() string {
	return r.rank
}

func (r *ordinalRank) Order() int {
	return r.gen.ranks[r.rank]
}

func (r *ordinalRank) All() []OrdinalRank {
	ranks := make([]OrdinalRank, 0, len(r.gen.ranks))
	for rank, _ := range r.gen.ranks {
		ranks = append(ranks, r.gen.Ord(rank))
	}

	return ranks
}

func (r *ordinalRank) Related(rank OrdinalRank) bool {
	other, ok := rank.(*ordinalRank)
	return ok && r.gen == other.gen
}

type nominalStat struct {
	items map[string]int
}

func newNominalStat() *nominalStat {
	n := new(nominalStat)
	n.items = make(map[string]int)
	return n
}

func (n *nominalStat) Item(item string) {
	c, ok := n.items[item]
	if !ok {
		n.items[item] = 1
	} else {
		n.items[item] = c + 1
	}
}

func (n *nominalStat) Init(items []string) {
	for _, item := range items {
		if _, ok := n.items[item]; !ok {
			n.items[item] = 0
		}
	}
}

func (n *nominalStat) combine(v *nominalStat) {
	for name, item := range v.items {
		a, ok := n.items[name]
		if !ok {
			a = 0
		}

		n.items[name] = a + item
	}
}

func (n *nominalStat) report() *NominalReport {
	items := make([]*NominalReportItem, 0, len(n.items))
	names := make([]string, 0, len(n.items))
	for name, _ := range n.items {
		names = append(names, name)
	}

	sort.Strings(names)
	count := 0
	for _, name := range names {
		c := n.items[name]
		count += c
		item := new(NominalReportItem)
		item.Name = name
		item.N = c
		items = append(items, item)
	}

	for _, item := range items {
		item.Percent = float32(item.N) * 100.0 / float32(count)
	}

	return &NominalReport{items}
}

type ordinalCount struct {
	rank OrdinalRank
	c    int
}

type ordinalStat struct {
	ranks map[string]*ordinalCount
}

func newOrdinalStat(ordinals []OrdinalRank) *ordinalStat {
	ord := new(ordinalStat)
	ord.ranks = make(map[string]*ordinalCount)
	for _, o := range ordinals {
		ord.ranks[o.Name()] = &ordinalCount{o, 0}
	}

	return ord
}

func (o *ordinalStat) Rank(rank OrdinalRank) {
	ord, ok := o.ranks[rank.Name()]
	if ok {
		ord.c++
	}
}

func (o *ordinalStat) combine(v *ordinalStat) {
	for name, rank := range v.ranks {
		a, ok := o.ranks[name]
		if !ok || !a.rank.Related(rank.rank) {
			return
		}

		a.c += rank.c
	}
}

func (o *ordinalStat) report() *OrdinalReport {
	items := make([]*OrdinalReportRank, len(o.ranks))
	if len(o.ranks) <= 0 {
		return &OrdinalReport{make([]*OrdinalReportRank, 0)}
	}

	count := 0
	for _, rank := range o.ranks {
		r := new(OrdinalReportRank)
		r.Name = rank.rank.Name()
		r.Order = rank.rank.Order()
		r.N = rank.c
		items[rank.rank.Order()] = r

		count += rank.c
	}

	if count == 0 {
		return &OrdinalReport{make([]*OrdinalReportRank, 0)}
	}

	up := 0
	down := count
	for _, item := range items {
		item.Percent = float32(item.N) * 100.0 / float32(count)

		up += item.N
		item.CumulativeN = up
		item.CumulativePercent = float32(up) * 100.0 / float32(count)

		item.DownCumulativeN = down
		item.DownCumulativePercent = float32(down) * 100.0 / float32(count)
		down -= item.N
	}

	return &OrdinalReport{items}
}

type intervalStat struct {
	values   []float64 // float64 不太好按 interval map……
	interval *float64
}

func newIntervalStat() *intervalStat {
	i := new(intervalStat)
	i.values = make([]float64, 0)
	return i
}

func (i *intervalStat) Value(value float64) {
	i.values = append(i.values, value)
}

func (i *intervalStat) Init(interval float64) {
	v := interval
	i.interval = &v
}

func (i *intervalStat) combine(v *intervalStat) {
	i.values = append(i.values, v.values...)
	if i.interval == nil && v.interval != nil {
		i.interval = v.interval
	}
}

func (i *intervalStat) report() *IntervalReport {
	if len(i.values) <= 0 {
		return &IntervalReport{1, make([]*IntervalReportItem, 0), 0, 0}
	}

	// TODO: define value's significant digit
	interval, min := calcInterval(i.values)
	if i.interval != nil {
		m := math.Mod(interval, *i.interval) / *i.interval
		if m > 1e-14 && m < 1-1e-14 {
			fmt.Println("set interval ", *i.interval, " can't div calc interval ", interval)
		}

		interval = *i.interval
	}

	steps := make(map[float64]int)
	stepValues := make(map[float64]float64)
	stepItems := make(map[float64][]float64)
	for _, v := range i.values {
		step, value := stepInterval(v, min, interval)
		c, ok := steps[step]
		if ok {
			steps[step] = c + 1
		} else {
			steps[step] = 1
			stepValues[step] = value
			stepItems[step] = make([]float64, 0)
		}

		stepItems[step] = append(stepItems[step], v)
	}

	s := make([]float64, 0, len(steps))
	for step, _ := range steps {
		s = append(s, step)
	}

	sort.Float64s(s)
	count := 0
	items := make([]*IntervalReportItem, 0, len(s))
	for _, step := range s {
		c := steps[step]
		item := new(IntervalReportItem)
		item.Value = stepValues[step]
		item.Step = int(step)
		item.N = c
		item.Mean = calcMean(stepItems[step])
		item.StandardDeviation = calcStandardDeviation(stepItems[step])
		count += c

		items = append(items, item)
	}

	up := 0
	down := count
	for _, item := range items {
		item.Percent = float32(item.N) * 100.0 / float32(count)

		up += item.N
		item.CumulativeN = up
		item.CumulativePercent = float32(up) * 100.0 / float32(count)

		item.DownCumulativeN = down
		item.DownCumulativePercent = float32(down) * 100.0 / float32(count)
		down -= item.N
	}

	var mean float64 = calcMean(i.values)
	var standardDeviation float64 = calcStandardDeviation(i.values)

	return &IntervalReport{interval, items, mean, standardDeviation}
}

type ratioStat struct {
	values []float64
}

func newRatioStat() *ratioStat {
	r := new(ratioStat)
	r.values = make([]float64, 0)
	return r
}

func (r *ratioStat) Value(value float64) {
	r.values = append(r.values, value)
}

func (r *ratioStat) combine(v *ratioStat) {
	r.values = append(r.values, v.values...)
}

func (r *ratioStat) report() *RatioReport {
	if len(r.values) <= 0 {
		return &RatioReport{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	}

	ratio := new(RatioReport)

	ratio.N = len(r.values)

	ratio.Mean = calcMean(r.values)
	ratio.GeometricMean = calcGeometricMean(r.values)
	ratio.QuadraticMean = calcQuadraticMean(r.values)
	ratio.HarmonicMean = calcHarmonicMean(r.values)

	ratio.StandardDeviation = calcStandardDeviation(r.values)

	sort.Float64s(r.values)
	ratio.P05 = r.values[ratio.N*5/100]
	ratio.P25 = r.values[ratio.N*25/100]
	ratio.P50 = r.values[ratio.N*50/100]
	ratio.P75 = r.values[ratio.N*75/100]
	ratio.P95 = r.values[ratio.N*95/100]
	return ratio
}
