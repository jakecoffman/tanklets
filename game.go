package tanklets

import (
	"net"

	"github.com/jakecoffman/cp"
)

type PlayerID uint16

var (
	// client only
	Me    PlayerID
	State int

	// server only
	Players = map[PlayerID]*net.UDPAddr{} // represent players separately for, e.g. disconnects
	Lookup  = map[string]PlayerID{}       // look up Addr

	// both client and server (TODO: Server needs to sync some of this still)
	Tanks         = map[PlayerID]*Tank{}
	Space         *cp.Space
	Width, Height int
)

// Game state
const (
	GAME_WAITING = iota
	GAME_PLAYING
	GAME_DEAD
)

// Collision types
const (
	COLLISION_TYPE_BULLET = 1
)

// Collision categories
const (
	_ = iota
)

var PLAYER_MASK_BIT uint = 1 << 31

var PlayerFilter cp.ShapeFilter = cp.ShapeFilter{
	cp.NO_GROUP, PLAYER_MASK_BIT, PLAYER_MASK_BIT,
}
var NotPlayerFilter cp.ShapeFilter = cp.ShapeFilter{
	cp.NO_GROUP, ^PLAYER_MASK_BIT, ^PLAYER_MASK_BIT,
}

func NewGame(width, height float64) {
	// physics
	space := cp.NewSpace()

	sides := []cp.Vector{
		{0, 0}, {0, height},
		{width, 0}, {width, height},
		{0, 0}, {width, 0},
		{0, height}, {width, height},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(1)
		seg.SetFriction(1)
		seg.SetFilter(PlayerFilter)
	}

	handler := space.NewWildcardCollisionHandler(COLLISION_TYPE_BULLET)
	handler.PreSolveFunc = BulletPreSolve

	Width = int(width)
	Height = int(height)
	Space = space
}

func Update(dt float64) {
	for _, tank := range Tanks {
		tank.Update(dt)
	}

	for _, bullet := range Bullets {
		bullet.Update(dt)
	}

	Space.Step(dt)
}
