package server

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/gutils"
	"github.com/jakecoffman/tanklets/pkt"
)

type Game struct {
	*tanklets.Game
	Network *Server
}

func NewGame(width, height float64, network *Server) *Game {
	game := tanklets.NewGame(width, height)

	fmt.Println("Server making some boxes")
	w, h := int(width), int(height)
	boxIdCursor := gutils.NewCursor(0, 1e9)
	for i := 10; i < h; i += 50 {
		box := game.NewBox(tanklets.BoxID(boxIdCursor.Next()))
		box.SetPosition(cp.Vector{X: width / 2, Y: float64(i)})
	}
	for i := 10; i < w; i += 50 {
		box := game.NewBox(tanklets.BoxID(boxIdCursor.Next()))
		box.SetPosition(cp.Vector{X: float64(i), Y: height / 2})
	}

	return &Game{
		Game:    game,
		Network: network,
	}
}

func (g *Game) Update(dt float64) {
	g.Game.Update(dt)

	switch {
	case g.Game.State == tanklets.StateStartCountdown:
		if time.Now().Sub(g.StartTime) > 3*time.Second {
			g.Game.State = tanklets.StatePlaying
			Players.SendAll(g.Network, pkt.State{State: tanklets.StatePlaying})
		}
	case g.Game.State > tanklets.StatePlaying:
		if time.Now().Sub(g.EndTime) > 3*time.Second {
			g.Game.State = tanklets.StateStartCountdown
			g.Game.StartTime = time.Now()
			Players.SendAll(g.Network, pkt.State{State: tanklets.StateStartCountdown})
			g.Restart()
		}
	}
}

func (g *Game) Restart() {
	for _, t := range g.Tanks {
		t.Destroyed = false
		t.SetPosition(cp.Vector{
			X: 10 + float64(rand.Intn(int(g.Width)-20)),
			Y: 10 + float64(rand.Intn(int(g.Height)-20)),
		})
		t.SetVelocityVector(cp.Vector{})
		t.SetAngularVelocity(0)
		t.SetAngle(0)
		t.NextMove = pkt.Move{}
		Players.SendAll(g.Network, t.Location())
	}
	w, h := int(g.Width), int(g.Height)
	var positions []cp.Vector
	for i := 10; i < h; i += 50 {
		positions = append(positions, cp.Vector{X: g.Width / 2, Y: float64(i)})
	}
	for i := 10; i < w; i += 50 {
		positions = append(positions, cp.Vector{X: float64(i), Y: g.Height / 2})
	}
	for i, b := range g.Boxes {
		b.SetPosition(positions[i])
		b.SetAngle(0)
		b.SetAngularVelocity(0)
		b.SetVelocityVector(cp.Vector{})
		Players.SendAll(g.Network, b.Location())
	}
	for _, b := range g.Bullets {
		b.Bounce = 100
		Players.SendAll(g.Network, b.Location())
		b.Destroy(true)
	}
}
