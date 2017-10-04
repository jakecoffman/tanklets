package tanklets

import (
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets/core"
)

const (
	tankw = 30
	tankh = 20

	turretw = 4
	turreth = 15

	turnSpeed = 1.0
	maxSpeed  = 60

	shotCooldown = 250 * time.Millisecond
)

type Tank struct {
	Turret
	*cp.Body
	*cp.Shape

	ControlBody *cp.Body
	LastShot    time.Time
}

type Turret struct {
	*cp.Body
	*cp.Shape
}

func NewTank(space *cp.Space) *Tank {
	tank := &Tank{}
	tank.ControlBody = space.AddBody(cp.NewKinematicBody())
	tank.Body = space.AddBody(cp.NewBody(1, cp.MomentForBox(1, tankw, tankh)))
	tankShape := space.AddShape(cp.NewBox(tank.Body, tankw, tankh, 2))
	tankShape.SetElasticity(2)
	tankShape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	pivot := space.AddConstraint(cp.NewPivotJoint2(tank.ControlBody, tank.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)
	pivot.SetMaxForce(10000)

	gear := space.AddConstraint(cp.NewGearJoint(tank.ControlBody, tank.Body, 0.0, 1.0))
	//gear.SetErrorBias(0) // attempt to fully correct the joint each step
	gear.SetMaxBias(5)
	gear.SetMaxForce(50000)

	tank.Turret.Body = space.AddBody(cp.NewKinematicBody())
	tank.Turret.Shape = space.AddShape(cp.NewSegment(tank.Turret.Body, cp.Vector{0, 0}, cp.Vector{turreth, 0}, turretw))
	tank.Turret.Shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
	circlePart := space.AddShape(cp.NewCircle(tank.Turret.Body, 10, cp.Vector{}))
	circlePart.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	return tank
}

func (tank *Tank) Update(space *cp.Space) {
	// update body
	tank.ControlBody.SetAngle(tank.Body.Angle() - core.Keyboard.X*turnSpeed)
	if core.Keyboard.Y == 0 {
		tank.ControlBody.SetVelocity(0, 0)
	} else {
		tank.ControlBody.SetVelocityVector(tank.Body.Rotation().Rotate(cp.Vector{maxSpeed * core.Keyboard.Y, 0.0}))
	}

	// update turret
	tank.Turret.SetPosition(tank.Body.Position())
	mouseDelta := core.Mouse.Sub(tank.Turret.Body.Position())
	turn := tank.Turret.Rotation().Unrotate(mouseDelta).ToAngle()
	tank.Turret.SetAngle(tank.Turret.Angle() - turn)

	// fire
	if core.RightClick && time.Now().Sub(tank.LastShot) > shotCooldown {
		tank.Shoot(space)
		tank.LastShot = time.Now()
	}
}

func (tank *Tank) Shoot(space *cp.Space) {
	bullet := NewBullet()
	pos := tank.Turret.Position()
	pos.X += 10
	bullet.Body.SetPosition(pos)
	bullet.Body.SetAngle(tank.Turret.Angle())
	bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{bulletSpeed, 0}))
	bullet.Shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
	space.AddBody(bullet.Body)
	space.AddShape(bullet.Shape)
}
