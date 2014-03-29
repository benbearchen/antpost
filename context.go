package antpost

import (
	"fmt"
	"sort"
	"time"
)

type Context struct {
	cur     *droneContext
	history []*droneContext
	count   int
	timer   *time.Timer
}

func NewContext() *Context {
	c := new(Context)
	c.history = make([]*droneContext, 0)
	c.count = -1
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
	}
}

func (c *Context) Report() []*Report {
	n := len(c.history)
	if n <= 0 {
		return make([]*Report, 0)
	}

	start := c.history[0].start
	end := c.history[0].end

	d := make([]int, 0, n)
	for _, h := range c.history {
		if h.start.Before(start) {
			start = h.start
		}

		if end.After(h.end) {
			end = h.end
		}

		d = append(d, int(h.end.Sub(h.start)/time.Millisecond))
	}

	sort.Sort(sort.IntSlice(d))

	var sum int64 = 0
	for _, d := range d {
		sum += int64(d)
	}

	p05 := time.Duration(d[n*5/100]) * time.Millisecond
	p50 := time.Duration(d[n*50/100]) * time.Millisecond
	p95 := time.Duration(d[n*95/100]) * time.Millisecond

	r := new(Report)
	r.Time.N = n
	r.Time.Avg = time.Millisecond * time.Duration(sum/int64(n))
	r.Time.P05 = p05
	r.Time.P50 = p50
	r.Time.P95 = p95

	return []*Report{r}
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
