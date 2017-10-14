package tanklets

import (
	"time"

	"github.com/jakecoffman/cp"
)

type PlayerID uint8

var (
	Me    PlayerID
	Tanks = map[PlayerID]*Tank{}
	// Server only lookup of Addr to ID
	Lookup = map[string]PlayerID{}

	Space *cp.Space = cp.NewSpace()

	State int

	Width, Height int

	IsClient    bool
	IsConnected bool
)

// Game state
const (
	GAME_START = iota
	GAME_ACTIVE
	GAME_MENU
	GAME_WIN
)

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
		//seg.SetFilter(examples.NotGrabbableFilter)
	}

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
			Bullets = append(Bullets[:i], Bullets[i+1:]...)
		} else {
			break
		}
	}

	Space.Step(dt)
}
