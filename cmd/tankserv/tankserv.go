package main

import (
	"log"
	"time"

	"github.com/jakecoffman/tanklets"
)

const (
	step         = 16666666
	stepDuration = step * time.Nanosecond
)

const updateRate = 2
var updateCount int

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tanklets.NewGame(800, 600)
	tanklets.IsServer = true
	tanklets.NetInit()
	defer func() { log.Println(tanklets.NetClose()) }()

	tick := time.Tick(stepDuration)
	var ticks int

	log.Println("Server Running")

	lastFrame := time.Now()
	var dt time.Duration

	ticklet := func() {
		currentFrame := time.Now()
		dt = currentFrame.Sub(lastFrame)
		lastFrame = currentFrame
		ticks++
		tanklets.Update(dt.Seconds())

		updateCount++
		if updateCount < updateRate {
			return
		}
		updateCount = 0

		for _, player := range tanklets.Players {
			for _, tank := range tanklets.Tanks {
				tanklets.Send(tank.Location(), player)
			}
		}
	}

	for {
		// ticks get priority, so try to tick first always
		select {
		case <-tick:
			ticklet()
		default:
		}

		// handle one incoming or one tick, whichever is next
		select {
		case <-tick:
			ticklet()
		case incoming := <-tanklets.Incomings:
			if err := incoming.Handler.Handle(incoming.Addr); err != nil {
				log.Fatal(err)
			}
		}
	}
}
