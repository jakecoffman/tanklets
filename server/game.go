package server

import (
	"math/rand"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
)

type Game struct {
	*tanklets.Game
}

func NewGame(width, height float64) *Game {
	return &Game{
		Game: tanklets.NewGame(width, height),
	}
}

func (g *Game) Update(dt float64) {
	g.Game.Update(dt)

	switch {
	case g.Game.State == tanklets.StateStartCountdown:
		if time.Now().Sub(g.StartTime) > 3*time.Second {
			g.Game.State = tanklets.StatePlaying
			Players.SendAll(pkt.State{State: tanklets.StatePlaying})
		}
	case g.Game.State > tanklets.StatePlaying:
		if time.Now().Sub(g.EndTime) > 3*time.Second {
			g.Game.State = tanklets.StateStartCountdown
			g.Game.StartTime = time.Now()
			Players.SendAll(pkt.State{State: tanklets.StateStartCountdown})
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
		Players.SendAll(t.Location())
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
		Players.SendAll(b.Location())
	}
	for _, b := range g.Bullets {
		b.Bounce = 100
		Players.SendAll(b.Location())
		b.Destroy(true)
	}
}
