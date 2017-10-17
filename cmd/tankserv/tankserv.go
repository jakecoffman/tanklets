package main

import (
	"log"
	"time"

	"github.com/jakecoffman/tanklets"
)

const (
	step             = 16666666
	stepDuration     = step * time.Nanosecond
	serverUpdateRate = 200 * time.Millisecond
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tanklets.NewGame(800, 600)
	tanklets.IsServer = true
	tanklets.NetInit()
	defer func() { log.Println(tanklets.NetClose()) }()

	tick := time.Tick(stepDuration)
	var ticks int

	update := time.Tick(serverUpdateRate)

	log.Println("Server Running")

	lastFrame := time.Now()
	var dt time.Duration

	ticklet := func() {
		currentFrame := time.Now()
		dt = currentFrame.Sub(lastFrame)
		lastFrame = currentFrame
		ticks++
		tanklets.Update(dt.Seconds())

		select {
		case <-update:
			for _, tank := range tanklets.Tanks {
				tanklets.Send(tank.Location(), tank.Addr)
			}
		default:
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
