package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
	"github.com/jakecoffman/tanklets/server"
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

	pingTick := time.Tick(1*time.Second)
	go func() {
		for range pingTick {
			ping := pkt.Ping{T: time.Now()}
			server.Players.SendAll(ping)
		}
	}()

	for {
		game := server.NewGame(800, 600)
		game.BulletCollisionHandler.PreSolveFunc = server.BulletPreSolve

		fmt.Println("Waiting for players to connect")

		for {
			// TODO: Move this above game creation, handle clients connecting, wait for people to
			//       connect from the lobby instead. Once the game starts, then start the countdown.
			select {
			case incoming := <-tanklets.IncomingPackets:
				server.ProcessNetwork(incoming, game)
			}

			allReady := true
			for _, t := range game.Tanks {
				if !t.Ready {
					allReady = false
					break
				}
			}
			if len(game.Tanks) > 0 && allReady {
				game.State = tanklets.GameStatePlaying
				server.Players.SendAll(pkt.State{State: tanklets.GameStatePlaying})
				break
			}
		}

		fmt.Println("Let's do this")
		Loop(game)
	}
}

var BoxLocations = map[tanklets.BoxID]pkt.BoxLocation{}

func Loop(game *server.Game) {
	physicsTick := time.Tick(time.Second / physicsTicks)
	updateTick := time.Tick(serverUpdates)
	var accumulator float64
	var dt time.Duration
	lastFrame := time.Now()

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

		if server.Players.Len() == 0 && server.HasHadPlayersConnect {
			fmt.Println("All players have disconnected, shutting down")
			server.HasHadPlayersConnect = false
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
				for _, tank := range game.Tanks {
					server.Players.SendAll(tank.Location())
				}
				for _, box := range game.Boxes {
					loc := box.Location()
					if loc != BoxLocations[box.ID] {
						server.Players.SendAll(loc)
						BoxLocations[box.ID] = loc
					}
				}
			}
		}
	}
}