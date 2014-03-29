package antpost

import (
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
	Time DurationReport
}
