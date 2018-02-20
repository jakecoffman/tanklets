package server

import (
	"github.com/jakecoffman/tanklets"
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
}