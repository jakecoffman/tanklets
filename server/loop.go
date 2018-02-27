package server

import (
	"fmt"
	"time"

	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
)

func Loop(network *Server) {
	pingTick := time.Tick(1*time.Second)
	go func() {
		for range pingTick {
			ping := pkt.Ping{T: time.Now()}
			Players.SendAll(network, ping)
		}
	}()

	game := NewGame(800, 600, network)
	game.BulletCollisionHandler.PreSolveFunc = BulletPreSolve
	game.BulletCollisionHandler.UserData = game

	fmt.Println("Waiting for players to connect")

	for {
		// TODO: Move this above game creation, handle clients connecting, wait for people to
		//       connect from the lobby instead. Once the game starts, then start the countdown.
		select {
		case incoming := <-network.IncomingPackets:
			ProcessNetwork(incoming, game)
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
			Players.SendAll(game.Network, pkt.State{State: tanklets.StateStartCountdown})
			break
		}
	}

	fmt.Println("Let's do this")
	// This seems hacky but it works
	time.Sleep(3*time.Second)
	game.State = tanklets.StatePlaying
	Players.SendAll(game.Network, pkt.State{State: tanklets.StatePlaying})

	Play(game)
}

var BoxLocations = map[tanklets.BoxID]pkt.BoxLocation{}

const (
	serverUpdates   = time.Second / 21.0
	physicsTicks    = 180.0
	physicsTickrate = 1.0 / physicsTicks
)

func Play(game *Game) {
	physicsTick := time.NewTicker(time.Second / physicsTicks)
	updateTick := time.NewTicker(serverUpdates)
	defer physicsTick.Stop()
	defer updateTick.Stop()

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

		if Players.Len() == 0 && HasHadPlayersConnect {
			fmt.Println("All players have disconnected, shutting down")
			HasHadPlayersConnect = false
			return
		}

		// handle all incoming messages this frame
	inner:
		for {
			select {
			case incoming := <-game.Network.IncomingPackets:
				ProcessNetwork(incoming, game)
			case <-physicsTick.C:
				// time to do a physics tick
				break inner
			case <-updateTick.C:
				for _, tank := range game.Tanks {
					Players.SendAll(game.Network, tank.Location())
				}
				for _, box := range game.Boxes {
					loc := box.Location()
					if loc != BoxLocations[box.ID] {
						Players.SendAll(game.Network, loc)
						BoxLocations[box.ID] = loc
					}
				}
			}
		}
	}
}
