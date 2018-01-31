package tanklets

import (
	"github.com/jakecoffman/cp"
	"github.com/engoengine/math"
	"github.com/jakecoffman/tanklets/gutils"
)

type PlayerID uint16

var (
	// TODO move this into client code... I think I can refactor the handlers to be in their respective packages
	// client only
	Me    PlayerID
)

// Game state
const (
	GameStateWaiting   = iota
	GameStatePlaying
	GameStateDead
)

// Collision types
const (
	CollisionTypeBullet = 1
)

var PlayerMaskBit uint = 1 << 31

var PlayerFilter = cp.ShapeFilter{
	cp.NO_GROUP, PlayerMaskBit, PlayerMaskBit,
}

var NotPlayerFilter = cp.ShapeFilter{
	cp.NO_GROUP, ^PlayerMaskBit, ^PlayerMaskBit,
}

type Game struct {
	Space   *cp.Space
	Bullets map[BulletID]*Bullet
	Tanks   map[PlayerID]*Tank

	Box *cp.Body

	State int

	playerIdCursor, color, bullet *gutils.Cursor
}

func NewGame(width, height float64) *Game {
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
		seg := space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(1)
		seg.SetFriction(0)
		seg.SetFilter(PlayerFilter)
	}

	const boxSize = 25
	Box := space.AddBody(cp.NewBody(1, cp.MomentForBox(1, boxSize, boxSize)))
	boxShape := space.AddShape(cp.NewBox(Box, boxSize, boxSize, 0))
	Box.SetPosition(cp.Vector{150, 150})
	boxShape.SetFriction(1)

	pivot := space.AddConstraint(cp.NewPivotJoint2(space.StaticBody, Box, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)       // disable joint correction
	pivot.SetMaxForce(1000.0) // emulate linear friction

	gear := space.AddConstraint(cp.NewGearJoint(space.StaticBody, Box, 0.0, 1.0))
	gear.SetMaxBias(0)
	gear.SetMaxForce(5000.0) // emulate angular friction

	handler := space.NewWildcardCollisionHandler(CollisionTypeBullet)
	handler.PreSolveFunc = BulletPreSolve

	return &Game{
		Space:   space,
		Bullets: map[BulletID]*Bullet{},
		Tanks:   map[PlayerID]*Tank{},

		Box: Box,

		// various cursors
		playerIdCursor: gutils.NewCursor(1, 100),
		color:          gutils.NewCursor(0, 14),
		bullet:         gutils.NewCursor(1, math.MaxInt64),
	}
}

func (g *Game) Update(dt float64) {
	for _, tank := range g.Tanks {
		tank.Update(dt)
	}

	for _, bullet := range g.Bullets {
		bullet.Update(dt)
	}
}
