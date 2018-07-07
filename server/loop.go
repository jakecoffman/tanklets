package server

import (
	"fmt"
	"time"

	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
)

func Lobby(network *Server) {
	pingTick := time.NewTicker(1 * time.Second)
	timeoutTick := time.NewTicker(5 * time.Second)
	defer pingTick.Stop()
	defer timeoutTick.Stop()

	done := make(chan struct{})
	defer func(){done<-struct{}{}}()
	go func() {
		for {
			select {
			case <-pingTick.C:
				network.Players.SendAll(network, pkt.Ping{T: time.Now()})
			case <-done:
				close(done)
				return
			}
		}
	}()

	game := NewGame(800, 600, network)
	game.BulletCollisionHandler.PreSolveFunc = BulletPreSolve
	game.BulletCollisionHandler.UserData = game

	fmt.Println("Waiting for players to connect")

	for {
	lobby:
		// TODO: Move this above game creation, handle clients connecting, wait for people to
		//       connect from the lobby instead. Once the game starts, then start the countdown.
		select {
		case incoming := <-network.IncomingPackets:
			ProcessNetwork(incoming, game)
		case <-timeoutTick.C:
			for id, tank := range game.Tanks {
				if time.Now().Sub(tank.LastPkt) > 10 * time.Second {
					delete(game.Tanks, id)
					network.Players.Delete(id)
					if len(game.Tanks) == 0 {
						return
					}
				}
			}
			goto lobby
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
			network.Players.SendAll(game.Network, pkt.State{State: tanklets.StateStartCountdown})
			network.Players.SendAll(game.Network, pkt.State{State: tanklets.StateStartCountdown})
			network.Players.SendAll(game.Network, pkt.State{State: tanklets.StateStartCountdown})
			break
		}
	}

	// one last scrub of disconnected players
	for id, tank := range game.Tanks {
		if time.Now().Sub(tank.LastPkt) > 10 * time.Second {
			delete(game.Tanks, id)
			network.Players.Delete(id)
		}
	}

	fmt.Println("Let's do this")
	// This seems hacky but it works
	time.Sleep(3 * time.Second)
	game.State = tanklets.StatePlaying
	network.Players.SendAll(game.Network, pkt.State{State: tanklets.StatePlaying})

	Play(game)
}

var BoxLocations = map[tanklets.BoxID]pkt.BoxLocation{}

const (
	netTickrate     = 20
	boxTickrate     = 10
	playerUpdates   = time.Second / netTickrate
	boxUpdates      = time.Second / boxTickrate
	physicsTicks    = 180.0
	physicsTickrate = 1.0 / physicsTicks
)

func Play(game *Game) {
	ticker := time.NewTicker(time.Second / physicsTicks)
	defer ticker.Stop()

	var lastUpdateTick, lastBoxTick time.Time

	var accumulator float64
	var dt time.Duration
	lastFrame := time.Now()

	for {
	netloop:
		for {
			select {
			case incoming := <-game.Network.IncomingPackets:
				ProcessNetwork(incoming, game)
			case <-ticker.C:
				break netloop
			}
		}

		currentFrame := time.Now()
		dt = currentFrame.Sub(lastFrame)
		lastFrame = currentFrame
		accumulator += dt.Seconds()

		if accumulator >= physicsTickrate {
			for _, tank := range game.Tanks {
				tank.FixedUpdate()
			}
			game.Space.Step(physicsTickrate)
			accumulator -= physicsTickrate
		}
		game.Update(dt.Seconds())

		if game.Network.Players.Len() == 0 && HasHadPlayersConnect {
			fmt.Println("All players have disconnected, shutting down")
			HasHadPlayersConnect = false
			return
		}

		if currentFrame.Sub(lastUpdateTick) >= playerUpdates {
			lastUpdateTick = currentFrame
			for _, tank := range game.Tanks {
				game.Network.Players.SendAll(game.Network, tank.Location())
			}
		}

		if currentFrame.Sub(lastBoxTick) >= boxUpdates {
			lastBoxTick = currentFrame
			for _, box := range game.Boxes {
				loc := box.Location()
				if loc != BoxLocations[box.ID] {
					game.Network.Players.SendAll(game.Network, loc)
					BoxLocations[box.ID] = loc
				}
			}
		}
	}
}
