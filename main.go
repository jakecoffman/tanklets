package tanklets

import (
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets/core"
)

const (
	width  = 640
	height = 480

	hwidth  = width / 2
	hheight = height / 2
)

var (
	tank1 *Tank
)

func Main() {
	space := cp.NewSpace()

	sides := []cp.Vector{
		{-hwidth, -hheight}, {-hwidth, hheight},
		{hwidth, -hheight}, {hwidth, hheight},
		{-hwidth, -hheight}, {hwidth, -hheight},
		{-hwidth, hheight}, {hwidth, hheight},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(1)
		seg.SetFriction(0)
		seg.SetFilter(core.NotGrabbableFilter)
	}

	tank1 = NewTank(space)
	tank1.Body.SetPosition(cp.Vector{-100, -100})

	core.Main(space, 1.0/60.0, update, core.DefaultDraw)
}

func update(space *cp.Space, dt float64) {
	space.Step(dt)
	tank1.Update(space)
}
