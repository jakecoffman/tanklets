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
	network := server.NewServer("0.0.0.0:1234")
	go network.Recv()
	defer func() { fmt.Println(network.Close()) }()

	fmt.Println("Server Running")

	pingTick := time.Tick(1*time.Second)
	go func() {
		for range pingTick {
			ping := pkt.Ping{T: time.Now()}
			server.Players.SendAll(network, ping)
		}
	}()

	for {
		game := server.NewGame(800, 600, network)
		game.BulletCollisionHandler.PreSolveFunc = server.BulletPreSolve
		game.BulletCollisionHandler.UserData = game

		fmt.Println("Waiting for players to connect")

		for {
			// TODO: Move this above game creation, handle clients connecting, wait for people to
			//       connect from the lobby instead. Once the game starts, then start the countdown.
			select {
			case incoming := <-network.IncomingPackets:
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
				game.State = tanklets.StateStartCountdown
				server.Players.SendAll(game.Network, pkt.State{State: tanklets.StateStartCountdown})
				break
			}
		}

		fmt.Println("Let's do this")
		// This seems hacky but it works
		time.Sleep(3*time.Second)
		game.State = tanklets.StatePlaying
		server.Players.SendAll(game.Network, pkt.State{State: tanklets.StatePlaying})

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
			case incoming := <-game.Network.IncomingPackets:
				server.ProcessNetwork(incoming, game)
			case <-physicsTick:
				// time to do a physics tick
				break inner
			case <-updateTick:
				for _, tank := range game.Tanks {
					server.Players.SendAll(game.Network, tank.Location())
				}
				for _, box := range game.Boxes {
					loc := box.Location()
					if loc != BoxLocations[box.ID] {
						server.Players.SendAll(game.Network, loc)
						BoxLocations[box.ID] = loc
					}
				}
			}
		}
	}
}