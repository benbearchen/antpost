package antpost

import (
	"sync"
	"time"
)

func Run(drone Drone, goroutines int, count int, d time.Duration) *Context {
	if goroutines <= 0 {
		return nil
	}

	contexts := make([]*Context, 0, goroutines)
	wg := new(sync.WaitGroup)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		context := NewContext()
		if count > 0 {
			context.SetCount(count)
		}

		if d > 0 {
			context.SetTime(d)
		}

		contexts = append(contexts, context)
		go func(d Drone, c *Context) {
			defer wg.Done()
			run(d, c)
		}(drone.Next(), context)
	}

	wg.Wait()
	contexts[0].Combine(contexts[1:]...)
	return contexts[0]
}

func run(drone Drone, context *Context) {
	for drone != nil {
		if !context.Start() {
			return
		}

		result := drone.Run(context)
		context.End(result)

		drone = drone.Next()
	}
}
