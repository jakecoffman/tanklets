package tanklets

import (
	"net"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

const (
	TurretWidth  = 4
	TurretHeight = 15

	TurnSpeed = .05
	MaxSpeed  = 60

	ShotCooldown = 250 * time.Millisecond
)

type Tank struct {
	// network
	ID   PlayerID
	Addr *net.UDPAddr

	// physics
	Turret
	*cp.Body
	*cp.Shape
	ControlBody *cp.Body

	width, height float64
	Color         mgl32.Vec3

	LastShot time.Time
}

type Turret struct {
	*cp.Body
	*cp.Shape

	width, height float32
}

const (
	tankWidth  = 20
	tankHeight = 30
)

func NewTank(id PlayerID, color mgl32.Vec3) *Tank {
	tank := &Tank{
		ID: id,
		Color: color,
	}
	tank.ControlBody = Space.AddBody(cp.NewKinematicBody())
	tank.Body = Space.AddBody(cp.NewBody(1, cp.MomentForBox(1, tankWidth, tankHeight)))
	tankShape := Space.AddShape(cp.NewBox(tank.Body, tankWidth, tankHeight, 2))
	tankShape.SetElasticity(0)
	tankShape.SetFriction(0)
	tankShape.SetFilter(cp.NewShapeFilter(uint(id), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	pivot := Space.AddConstraint(cp.NewPivotJoint2(tank.ControlBody, tank.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)
	pivot.SetMaxForce(10000)

	gear := Space.AddConstraint(cp.NewGearJoint(tank.ControlBody, tank.Body, 0.0, 1.0))
	gear.SetErrorBias(0) // attempt to fully correct the joint each step
	gear.SetMaxBias(5)
	gear.SetMaxForce(50000)

	tank.Turret.Body = Space.AddBody(cp.NewKinematicBody())
	tank.Turret.Shape = Space.AddShape(cp.NewSegment(tank.Turret.Body, cp.Vector{0, 0}, cp.Vector{TurretHeight, 0}, TurretWidth))
	tank.Turret.Shape.SetFilter(cp.NewShapeFilter(uint(id), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
	circlePart := Space.AddShape(cp.NewCircle(tank.Turret.Body, 10, cp.Vector{}))
	circlePart.SetFilter(cp.NewShapeFilter(uint(id), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	return tank
}

func (tank *Tank) Update() {
	// update body

	// update turret
	tank.Turret.SetPosition(tank.Body.Position())
}

func (tank *Tank) Shoot(space *cp.Space) {
	bullet := NewBullet(tank.Color)

	pos := cp.Vector{X: tankHeight / 2.0}
	pos = pos.Rotate(tank.Turret.Rotation())
	bullet.Body.SetPosition(pos.Add(tank.Turret.Position()))
	bullet.Body.SetAngle(tank.Turret.Angle())
	bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{bulletSpeed, 0}))
	bullet.Shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	space.AddBody(bullet.Body)
	space.AddShape(bullet.Shape)
}

// gather important data to transmit
func (tank *Tank) Location() *Location {
	return &Location{
		ID:              tank.ID,
		X:               tank.Body.Position().X,
		Y:               tank.Body.Position().Y,
		Angle:           tank.Body.Angle(),
		AngularVelocity: tank.Body.AngularVelocity(),
		Vx:              tank.Body.Velocity().X,
		Vy:              tank.Body.Velocity().Y,

		Turret: tank.Turret.Angle(),
	}
}
