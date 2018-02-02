package tanklets

import (
	"github.com/jakecoffman/cp"
	"github.com/engoengine/math"
	"github.com/jakecoffman/tanklets/gutils"
	"math/rand"
	"fmt"
)

type PlayerID uint16

var (
	// TODO move this into client code... I think I can refactor the handlers to be in their respective packages
	// client only
	Me PlayerID
)

// Game state
const (
	GameStateWaiting = iota
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

type BoxID uint16

type Game struct {
	Space *cp.Space

	Bullets map[BulletID]*Bullet
	Tanks   map[PlayerID]*Tank
	Boxes   map[BoxID]*Box

	State int

	playerIdCursor, color, bullet *gutils.Cursor
}

func NewGame(width, height float64) *Game {
	// physics
	space := cp.NewSpace()

	game := &Game{
		Space:   space,
		Bullets: map[BulletID]*Bullet{},
		Tanks:   map[PlayerID]*Tank{},
		Boxes:   map[BoxID]*Box{},

		// various cursors
		playerIdCursor: gutils.NewCursor(1, 100),
		color:          gutils.NewCursor(0, 14),
		bullet:         gutils.NewCursor(1, math.MaxInt64),
	}

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

	if IsServer {
		fmt.Println("Server making some boxes")
		for i := 0; i < 100; i++ {
			box := game.NewBox(BoxID(i))
			box.SetPosition(cp.Vector{X: float64(rand.Intn(int(width-10))), Y: float64(rand.Intn(int(height-10)))})
		}
	}

	handler := space.NewWildcardCollisionHandler(CollisionTypeBullet)
	handler.PreSolveFunc = BulletPreSolve

	return game
}

func (g *Game) Update(dt float64) {
	for _, tank := range g.Tanks {
		tank.Update(dt)
	}

	for _, bullet := range g.Bullets {
		bullet.Update(dt)
	}
}
