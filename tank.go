package tanklets

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets/pkt"
)

// default tank attributes (power-ups could change them!)
const (
	TankWidth  = 20
	TankHeight = 32

	TurretWidth  = 4
	TurretHeight = 15

	TurnSpeed = 3
	MaxSpeed  = 60

	ShotCooldown = 333 * time.Millisecond
)

type Tank struct {
	// network
	ID PlayerID

	// physics
	Turret
	*cp.Body
	*cp.Shape
	ControlBody *cp.Body

	width, height float64
	Color         mgl32.Vec3
	Name          string
	Score         int

	LastShot time.Time

	Destroyed bool

	NextMove, LastMove pkt.Move
	LastPkt            time.Time
	Ready              bool
}

type Turret struct {
	*cp.Body
	*cp.Shape

	width, height float64
}

func (g *Game) NewTank(id PlayerID, color mgl32.Vec3) *Tank {
	tank := &Tank{
		ID:    id,
		Name:  fmt.Sprintf("Player %v", id),
		Color: color,
	}
	tank.Body = g.Space.AddBody(cp.NewBody(1, cp.MomentForBox(1, TankWidth, TankHeight)))
	tankShape := g.Space.AddShape(cp.NewBox(tank.Body, TankWidth, TankHeight, 0))
	tankShape.SetElasticity(0)
	tankShape.SetFriction(1)
	tankShape.SetFilter(cp.NewShapeFilter(uint(id), PlayerMaskBit, PlayerMaskBit))
	tankShape.UserData = tank

	tank.ControlBody = g.Space.AddBody(cp.NewKinematicBody())

	pivot := g.Space.AddConstraint(cp.NewPivotJoint2(tank.ControlBody, tank.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)      // prevent joint from sucking the tank in
	pivot.SetMaxForce(10000) // prevent tanks from spinning crazy

	g.Space.AddConstraint(cp.NewGearJoint(tank.ControlBody, tank.Body, 0.0, 1.0))
	//gear.SetErrorBias(0) // idk
	//gear.SetMaxBias(1.2) // idk
	//gear.SetMaxForce(50000) // don't set or tank will go through walls

	tank.Turret.Body = g.Space.AddBody(cp.NewKinematicBody())
	tank.Turret.Shape = g.Space.AddShape(cp.NewSegment(tank.Turret.Body, cp.Vector{0, 0}, cp.Vector{TurretHeight, 0}, TurretWidth))
	tank.Turret.Shape.SetFilter(cp.NewShapeFilter(uint(id), ^cp.ALL_CATEGORIES, ^cp.ALL_CATEGORIES))
	circlePart := g.Space.AddShape(cp.NewCircle(tank.Turret.Body, 10, cp.Vector{}))
	circlePart.SetFilter(cp.NewShapeFilter(uint(id), ^cp.ALL_CATEGORIES, ^cp.ALL_CATEGORIES))

	return tank
}

func (tank *Tank) Update(dt float64) {

}

func (tank *Tank) FixedUpdate(dt float64) {
	if tank.Destroyed {
		// slowly stop, looks cool
		tank.ControlBody.SetAngularVelocity(tank.ControlBody.AngularVelocity() * .99)
		tank.ControlBody.SetVelocityVector(tank.ControlBody.Velocity().Mult(.99))
		return
	}

	move := tank.NextMove
	if !(move.Turn == 0 && tank.LastMove.Turn == 0) {
		tank.ControlBody.SetAngularVelocity(float64(move.Turn) * TurnSpeed)
	}

	if !(move.Throttle == 0 && tank.LastMove.Throttle == 0) {
		tank.ControlBody.SetVelocityVector(tank.Body.Rotation().Rotate(cp.Vector{Y: float64(move.Throttle) * MaxSpeed}))
	}

	tank.Turret.SetPosition(tank.Body.Position())

	tank.LastMove = tank.NextMove
}

var locationSequence uint64

// gather important data to transmit
func (tank *Tank) Location() *pkt.Location {
	locationSequence++
	return &pkt.Location{
		ID:              tank.ID,
		Sequence:        locationSequence,
		X:               float32(tank.Body.Position().X),
		Y:               float32(tank.Body.Position().Y),
		Angle:           float32(tank.Body.Angle()),
		AngularVelocity: float32(tank.Body.AngularVelocity()),
		Vx:              float32(tank.Body.Velocity().X),
		Vy:              float32(tank.Body.Velocity().Y),

		Turret: float32(tank.Turret.Angle()),
	}
}
