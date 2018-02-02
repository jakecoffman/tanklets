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
	game := tanklets.NewGame(800, 600)
	tanklets.IsServer = true
	tanklets.NetInit("0.0.0.0:1234")
	defer func() { fmt.Println(tanklets.NetClose()) }()

	fmt.Println("Server Running")

	var hasHadPlayersConnect bool
	accumulator := 0.
	lastFrame := time.Now()
	var dt time.Duration

	physicsTick := time.Tick(time.Second / physicsTicks)
	updateTick := time.Tick(serverUpdates)
	pingTick := time.Tick(1*time.Second)
	go func() {
		for range pingTick {
			ping := tanklets.Ping{T: time.Now()}
			tanklets.Players.SendAll(ping)
		}
	}()

	for {
		currentFrame := time.Now()
		dt = currentFrame.Sub(lastFrame)
		lastFrame = currentFrame
		accumulator += dt.Seconds()

		if accumulator >= physicsTickrate {
			for _, tank := range game.Tanks {
				tank.FixedUpdate(physicsTickrate)
			}
			game.Space.Step(physicsTickrate)
			accumulator -= physicsTickrate
		}
		game.Update(dt.Seconds())

		// TODO move this check to the disconnect handler
		if !hasHadPlayersConnect && tanklets.Players.Len() > 0 {
			hasHadPlayersConnect = true
		}
		if tanklets.Players.Len() == 0 && hasHadPlayersConnect {
			fmt.Println("All players have disconnected, shutting down")
			return
		}

		// handle all incoming messages this frame
	inner:
		for {
			select {
			case incoming := <-tanklets.Incomings:
				incoming.Handler.Handle(incoming.Addr, game)
			case <-physicsTick:
				// time to do a physics tick
				break inner
			case <-updateTick:
				// 58 bytes per n players, 10 times per second = 580n^2
				for _, tank := range game.Tanks {
					tanklets.Players.SendAll(tank.Location())
				}
			}
		}
	}
}
