package antpost

import (
	"fmt"
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
	r.Time.Analyze(d)
	r.OKTime.Analyze(okd)

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
