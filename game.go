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
		// outer walls
		{0, 0}, {0, height},
		{width, 0}, {width, height},
		{0, 0}, {width, 0},
		{0, height}, {width, height},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(1)
		seg.SetFriction(0)
		seg.SetFilter(PlayerFilter)
	}

	const boxSize = 25
	boxBody := space.AddBody(cp.NewBody(1, cp.MomentForBox(1, boxSize, boxSize)))
	boxShape := space.AddShape(cp.NewBox(boxBody, boxSize, boxSize, 0))
	boxBody.SetPosition(cp.Vector{150, 150})
	boxShape.SetFriction(1)

	pivot := space.AddConstraint(cp.NewPivotJoint2(space.StaticBody, boxBody, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)       // disable joint correction
	pivot.SetMaxForce(1000.0) // emulate linear friction

	gear := space.AddConstraint(cp.NewGearJoint(space.StaticBody, boxBody, 0.0, 1.0))
	gear.SetMaxBias(0)
	gear.SetMaxForce(5000.0) // emulate angular friction

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

	for _, tank := range Tanks {
		tank.Turret.SetPosition(tank.Body.Position())
	}
}
