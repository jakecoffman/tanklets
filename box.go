package tanklets

import "github.com/jakecoffman/cp"

const boxSize = 25

type Box struct {
	*cp.Body

	ID BoxID
}

func (g *Game) NewBox(id BoxID) *Box {
	box := g.Space.AddBody(cp.NewBody(1, cp.MomentForBox(1, boxSize, boxSize)))
	boxShape := g.Space.AddShape(cp.NewBox(box, boxSize, boxSize, 0))
	box.SetPosition(cp.Vector{150, 150})
	boxShape.SetFriction(1)

	pivot := g.Space.AddConstraint(cp.NewPivotJoint2(g.Space.StaticBody, box, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)       // disable joint correction
	pivot.SetMaxForce(1000.0) // emulate linear friction

	gear := g.Space.AddConstraint(cp.NewGearJoint(g.Space.StaticBody, box, 0.0, 1.0))
	gear.SetMaxBias(0)
	gear.SetMaxForce(5000.0) // emulate angular friction

	g.Boxes[id] = &Box{Body: box, ID: id}
	return g.Boxes[id]
}

func (b *Box) Location() *BoxLocation {
	return &BoxLocation{
		ID:              b.ID,
		X:               float32(b.Body.Position().X),
		Y:               float32(b.Body.Position().Y),
		Angle:           float32(b.Body.Angle()),
	}
}
