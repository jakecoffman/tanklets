package tanklets

import (
	"time"

	"github.com/jakecoffman/cp"
)

type PlayerID uint16

var (
	Me    PlayerID
	Tanks = map[PlayerID]*Tank{}
	// Server only lookup of Addr to ID
	Lookup = map[string]PlayerID{}

	Space *cp.Space

	State int

	Width, Height int

	IsClient    bool
	IsConnected bool
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
		tank.Update()
	}

	now := time.Now()
	for i := 0; i < len(Bullets); {
		if now.Sub(Bullets[i].firedAt) > bulletTTL {
			Bullets[i].Destroy()
			Bullets = append(Bullets[:i], Bullets[i+1:]...)
		} else {
			break
		}
	}

	Space.Step(dt)
}
