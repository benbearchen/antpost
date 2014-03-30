package main

import (
	"github.com/benbearchen/antpost"
	"github.com/benbearchen/antpost/drones"

	"fmt"
	"time"
)

func main() {
	h := drones.NewHttpGetReq("http://localhost/about", nil, nil)
	d := drones.NewHttpDrone(h)
	for i := 1; i <= 256; i *= 2 {
		c := antpost.Run(d, i, 0, 15*time.Second)
		fmt.Println("goroutines: ", i, c.Report()[0])
		time.Sleep(time.Second * 1)
	}
}
