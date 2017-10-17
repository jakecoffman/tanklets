package tanklets

import (
	"net"
	"time"

	"log"

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

	Destroyed bool
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
		ID:    id,
		Color: color,
	}
	tank.ControlBody = Space.AddBody(cp.NewKinematicBody())
	tank.Body = Space.AddBody(cp.NewBody(1, cp.MomentForBox(1, tankWidth, tankHeight)))
	tankShape := Space.AddShape(cp.NewBox(tank.Body, tankWidth, tankHeight, 2))
	tankShape.SetElasticity(0)
	tankShape.SetFriction(0)
	tankShape.SetFilter(cp.NewShapeFilter(uint(id), PLAYER_MASK_BIT, PLAYER_MASK_BIT))
	tankShape.UserData = tank

	pivot := Space.AddConstraint(cp.NewPivotJoint2(tank.ControlBody, tank.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)
	pivot.SetMaxForce(10000)

	gear := Space.AddConstraint(cp.NewGearJoint(tank.ControlBody, tank.Body, 0.0, 1.0))
	//gear.SetErrorBias(0) // attempt to fully correct the joint each step
	gear.SetMaxBias(5)
	gear.SetMaxForce(50000)

	tank.Turret.Body = Space.AddBody(cp.NewKinematicBody())
	tank.Turret.Shape = Space.AddShape(cp.NewSegment(tank.Turret.Body, cp.Vector{0, 0}, cp.Vector{TurretHeight, 0}, TurretWidth))
	tank.Turret.Shape.SetFilter(cp.NewShapeFilter(uint(id), ^cp.ALL_CATEGORIES, ^cp.ALL_CATEGORIES))
	circlePart := Space.AddShape(cp.NewCircle(tank.Turret.Body, 10, cp.Vector{}))
	circlePart.SetFilter(cp.NewShapeFilter(uint(id), ^cp.ALL_CATEGORIES, ^cp.ALL_CATEGORIES))

	return tank
}

func (tank *Tank) Update() {
	// update body

	// update turret
	tank.Turret.SetPosition(tank.Body.Position())
}

func (tank *Tank) Shoot(space *cp.Space) {
	Send(Shoot{}, ServerAddr)
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

func (tank *Tank) Damage(bullet *Bullet) {
	if !IsServer {
		log.Println("I shouldn't be here...")
		return
	}

	if tank.Destroyed {
		return
	}

	tank.Destroyed = true

	log.Println("Tank", tank.ID, "destroyed by Tank", bullet.Tank.ID)

	for _, p := range Tanks {
		Send(Damage{tank.ID}, p.Addr)
	}
}
