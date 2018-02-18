package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jakecoffman/tanklets"
	"math/rand"
	"github.com/jakecoffman/tanklets/server"
	"github.com/jakecoffman/tanklets/pkt"
)

const (
	serverUpdates   = time.Second / 10.0
	physicsTicks    = 180.0
	physicsTickrate = 1.0 / physicsTicks
)

func main() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tanklets.IsServer = true
	tanklets.NetInit("0.0.0.0:1234")
	go server.Recv()
	defer func() { fmt.Println(tanklets.NetClose()) }()

	fmt.Println("Server Running")

	var hasHadPlayersConnect bool
	var accumulator float64
	var dt time.Duration
	lastFrame := time.Now()

	physicsTick := time.Tick(time.Second / physicsTicks)
	updateTick := time.Tick(serverUpdates)
	pingTick := time.Tick(1*time.Second)
	go func() {
		for range pingTick {
			ping := pkt.Ping{T: time.Now()}
			tanklets.Players.SendAll(ping)
		}
	}()

	game := tanklets.NewGame(800, 600)

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
			case incoming := <-tanklets.IncomingPackets:
				server.ProcessNetwork(incoming, game)
			case <-physicsTick:
				// time to do a physics tick
				break inner
			case <-updateTick:
				// 58 bytes per n players, 10 times per second = 580n^2
				for _, tank := range game.Tanks {
					tanklets.Players.SendAll(tank.Location())
				}
				for _, box := range game.Boxes {
					loc := box.Location()
					if loc != BoxLocations[box.ID] {
						tanklets.Players.SendAll(loc)
						BoxLocations[box.ID] = loc
					}
				}
			}
		}
	}
}

var BoxLocations = map[tanklets.BoxID]pkt.BoxLocation{}
