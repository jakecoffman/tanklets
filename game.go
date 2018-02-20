package tanklets

import (
	"fmt"
	"math"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets/gutils"
	"github.com/jakecoffman/tanklets/pkt"
)

type PlayerID = pkt.PlayerID
type BulletID = pkt.BulletID
type BoxID = pkt.BoxID

// Game state
const (
	GameStateWaiting = iota
	GameStatePlaying
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
	Space *cp.Space

	Bullets map[BulletID]*Bullet
	Tanks   map[PlayerID]*Tank
	Boxes   map[BoxID]*Box

	BulletCollisionHandler *cp.CollisionHandler

	Walls []*cp.Shape

	State int

	CursorPlayerId, CursorColor, CursorBullet *gutils.Cursor
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
		CursorPlayerId: gutils.NewCursor(1, 100),
		CursorColor:    gutils.NewCursor(0, 14),
		CursorBullet:   gutils.NewCursor(1, math.MaxInt64),
	}

	sides := []cp.Vector{
		// outer walls
		{0, 0}, {0, height},
		{width, 0}, {width, height},
		{0, 0}, {width, 0},
		{0, height}, {width, height},
	}

	for i := 0; i < len(sides); i += 2 {
		seg := space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 1))
		seg.SetElasticity(1)
		seg.SetFriction(0)
		seg.SetFilter(PlayerFilter)
		game.Walls = append(game.Walls, seg)
	}

	if IsServer {
		fmt.Println("Server making some boxes")
		w, h := int(width), int(height)
		boxIdCursor := gutils.NewCursor(0, 1e9)
		for i := 10; i<h; i += 50 {
			box := game.NewBox(BoxID(boxIdCursor.Next()))
			box.SetPosition(cp.Vector{X: width/2, Y: float64(i)})
		}
		for i := 10; i<w; i += 50 {
			box := game.NewBox(BoxID(boxIdCursor.Next()))
			box.SetPosition(cp.Vector{X: float64(i), Y: height/2})
		}
	}

	game.BulletCollisionHandler = space.NewWildcardCollisionHandler(CollisionTypeBullet)
	game.BulletCollisionHandler.PreSolveFunc = func(arb *cp.Arbiter, _ *cp.Space, _ interface{}) bool {
		arb.Ignore()
		return false
	}

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
