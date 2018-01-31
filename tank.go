package tanklets

import (
	"fmt"
	"log"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"math"
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

	LastShot time.Time

	Destroyed bool

	NextMove, LastMove Move
	Ping time.Duration // tank ping in ns
}

type Turret struct {
	*cp.Body
	*cp.Shape

	width, height float64
}

func NewTank(id PlayerID, color mgl32.Vec3) *Tank {
	tank := &Tank{
		ID:    id,
		Color: color,
	}
	tank.Body = Space.AddBody(cp.NewBody(1, cp.MomentForBox(1, TankWidth, TankHeight)))
	tankShape := Space.AddShape(cp.NewBox(tank.Body, TankWidth, TankHeight, 0))
	tankShape.SetElasticity(0)
	tankShape.SetFriction(1)
	tankShape.SetFilter(cp.NewShapeFilter(uint(id), PLAYER_MASK_BIT, PLAYER_MASK_BIT))
	tankShape.UserData = tank

	tank.ControlBody = Space.AddBody(cp.NewKinematicBody())

	pivot := Space.AddConstraint(cp.NewPivotJoint2(tank.ControlBody, tank.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)      // prevent joint from sucking the tank in
	pivot.SetMaxForce(10000) // prevent tanks from spinning crazy

	Space.AddConstraint(cp.NewGearJoint(tank.ControlBody, tank.Body, 0.0, 1.0))
	//gear.SetErrorBias(0) // idk
	//gear.SetMaxBias(1.2) // idk
	//gear.SetMaxForce(50000) // don't set or tank will go through walls

	tank.Turret.Body = Space.AddBody(cp.NewKinematicBody())
	tank.Turret.Shape = Space.AddShape(cp.NewSegment(tank.Turret.Body, cp.Vector{0, 0}, cp.Vector{TurretHeight, 0}, TurretWidth))
	tank.Turret.Shape.SetFilter(cp.NewShapeFilter(uint(id), ^cp.ALL_CATEGORIES, ^cp.ALL_CATEGORIES))
	circlePart := Space.AddShape(cp.NewCircle(tank.Turret.Body, 10, cp.Vector{}))
	circlePart.SetFilter(cp.NewShapeFilter(uint(id), ^cp.ALL_CATEGORIES, ^cp.ALL_CATEGORIES))

	return tank
}

func (tank *Tank) Update(dt float64) {

}

func (tank *Tank) FixedUpdate(dt float64) {
	if tank.Destroyed {
		return
	}

	m := tank.NextMove
	if !(m.Turn == 0 && tank.LastMove.Turn == 0) {
		tank.ControlBody.SetAngularVelocity(float64(m.Turn) * TurnSpeed)
	}

	if !(m.Throttle == 0 && tank.LastMove.Throttle == 0) {
		tank.ControlBody.SetVelocityVector(tank.Body.Rotation().Rotate(cp.Vector{Y: float64(m.Throttle) * MaxSpeed}))
	}

	x := math.Cos(m.TurretAngle)
	y := math.Sin(m.TurretAngle)
	angle := tank.Turret.Rotation().Unrotate(cp.Vector{x, y}).ToAngle()
	tank.Turret.SetAngle(tank.Turret.Angle() - angle)
	tank.Turret.SetPosition(tank.Body.Position())

	tank.LastMove = tank.NextMove
}

// gather important data to transmit
func (tank *Tank) Location() *Location {
	return &Location{
		ID:              tank.ID,
		X:               float32(tank.Body.Position().X),
		Y:               float32(tank.Body.Position().Y),
		Angle:           float32(tank.Body.Angle()),
		AngularVelocity: float32(tank.Body.AngularVelocity()),
		Vx:              float32(tank.Body.Velocity().X),
		Vy:              float32(tank.Body.Velocity().Y),

		Turret: float32(tank.Turret.Angle()),
	}
}

func (tank *Tank) Damage(bullet *Bullet) {
	if !IsServer {
		log.Println("I shouldn't be here...")
		return
	}

	if tank.Destroyed {
		return
	}

	tank.Destroyed = true
	tank.Body.SetVelocity(0, 0)
	tank.Body.SetAngularVelocity(0)

	if tank.ControlBody != nil {
		tank.ControlBody.SetVelocity(0, 0)
		tank.ControlBody.SetAngularVelocity(0)
	}

	fmt.Println("Tank", tank.ID, "destroyed by Tank", bullet.PlayerID, "bullet", bullet.ID)

	Players.SendAll(Damage{tank.ID})
}
