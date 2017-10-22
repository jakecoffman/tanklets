package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jakecoffman/tanklets"
)

const (
	step         = 16666666
	stepDuration = step * time.Nanosecond
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tanklets.NewGame(800, 600)
	tanklets.IsServer = true
	tanklets.NetInit()
	defer func() { fmt.Println(tanklets.NetClose()) }()

	tick := time.Tick(stepDuration)
	var ticks int

	fmt.Println("Server Running")

	lastFrame := time.Now()
	var dt time.Duration

	ticklet := func() {
		currentFrame := time.Now()
		dt = currentFrame.Sub(lastFrame)
		lastFrame = currentFrame
		ticks++
		tanklets.Update(dt.Seconds())

		for _, player := range tanklets.Players {
			for _, tank := range tanklets.Tanks {
				tanklets.Send(tank.Location(), player)
			}
		}
	}

	var hasHadPlayersConnect bool

	for {
		if len(tanklets.Players) > 0 {
			hasHadPlayersConnect = true
		}
		if len(tanklets.Players) == 0 && hasHadPlayersConnect {
			fmt.Println("All players have disconnected, shutting down")
			return
		}

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
			incoming.Handler.Handle(incoming.Addr)
		}
	}
}
