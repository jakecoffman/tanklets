package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jakecoffman/tanklets"
)

const (
	serverUpdates   = time.Second / 10.0
	physicsTicks    = 180.0
	physicsTickrate = 1.0 / physicsTicks
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tanklets.NewGame(800, 600)
	tanklets.IsServer = true
	tanklets.NetInit()
	defer func() { fmt.Println(tanklets.NetClose()) }()

	fmt.Println("Server Running")

	var hasHadPlayersConnect bool
	accumulator := 0.
	lastFrame := time.Now()
	var dt time.Duration

	physicsTick := time.Tick(time.Second / physicsTicks)
	updateTick := time.Tick(serverUpdates)

	for {
		currentFrame := time.Now()
		dt = currentFrame.Sub(lastFrame)
		lastFrame = currentFrame
		accumulator += dt.Seconds()

		if accumulator >= physicsTickrate {
			for _, tank := range tanklets.Tanks {
				tank.FixedUpdate(physicsTickrate)
			}
			tanklets.Space.Step(physicsTickrate)
			accumulator -= physicsTickrate
		}
		tanklets.Update(dt.Seconds())

		// TODO move this check to the disconnect handler
		if !hasHadPlayersConnect && len(tanklets.Players) > 0 {
			hasHadPlayersConnect = true
		}
		if len(tanklets.Players) == 0 && hasHadPlayersConnect {
			fmt.Println("All players have disconnected, shutting down")
			return
		}

		// handle all incoming messages this frame
	net:
		for {
			select {
			case incoming := <-tanklets.Incomings:
				incoming.Handler.Handle(incoming.Addr)
			case <-physicsTick:
				break net
			case <-updateTick:
				// 58 bytes per n players, 10 times per second = 580n^2
				for _, tank := range tanklets.Tanks {
					data, err := tank.Location().MarshalBinary()
					if err != nil {
						log.Println(err)
						continue
					}
					for _, player := range tanklets.Players {
						tanklets.SendRaw(data, player)
					}
				}
			}
		}
	}
}
