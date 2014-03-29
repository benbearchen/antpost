package antpost

type Drone interface {
	Run(context *Context) DroneResult
	Next() Drone
}
