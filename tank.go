package tanklets

import (
	"time"

	"github.com/jakecoffman/cp"
	"github.com/go-gl/mathgl/mgl32"
)

const (
	turretw = 4
	turreth = 15

	turnSpeed = .1
	maxSpeed  = 60

	shotCooldown = 250 * time.Millisecond
)

type Tank struct {
	Turret
	*cp.Body
	*cp.Shape

	texture *Texture2D
	width, height float64
	color mgl32.Vec3

	ControlBody *cp.Body
	LastShot    time.Time
}

type Turret struct {
	*cp.Body
	*cp.Shape
}

func NewTank(space *cp.Space, texture *Texture2D, w, h int) *Tank {
	width := float64(w)
	height := float64(h)
	tank := &Tank{
		width: width,
		height: height,
		texture: texture,
	}
	tank.ControlBody = space.AddBody(cp.NewKinematicBody())
	tank.Body = space.AddBody(cp.NewBody(1, cp.MomentForBox(1, width, height)))
	tankShape := space.AddShape(cp.NewBox(tank.Body, width, height, 2))
	tankShape.SetElasticity(2)
	tankShape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	pivot := space.AddConstraint(cp.NewPivotJoint2(tank.ControlBody, tank.Body, cp.Vector{}, cp.Vector{}))
	pivot.SetMaxBias(0)
	pivot.SetMaxForce(10000)

	gear := space.AddConstraint(cp.NewGearJoint(tank.ControlBody, tank.Body, 0.0, 1.0))
	gear.SetErrorBias(0) // attempt to fully correct the joint each step
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


	// update turret
	//tank.Turret.SetPosition(tank.Body.Position())
	//mouseDelta := core.Mouse.Sub(tank.Turret.Body.Position())
	//turn := tank.Turret.Rotation().Unrotate(mouseDelta).ToAngle()
	//tank.Turret.SetAngle(tank.Turret.Angle() - turn)
	//
	//// fire
	//if core.RightClick && time.Now().Sub(tank.LastShot) > shotCooldown {
	//	tank.Shoot(space)
	//	tank.LastShot = time.Now()
	//}
}

func (tank *Tank) Draw(renderer *SpriteRenderer) {
	pos := tank.Position()
	x, y := float32(pos.X - tank.width/2), float32(pos.Y - tank.height/2)
	renderer.DrawSprite(tank.texture, mgl32.Vec2{x, y}, tank.Size(), tank1.Angle(), tank.color)
}

func (tank *Tank) Size() mgl32.Vec2 {
	return mgl32.Vec2{float32(tank1.width), float32(tank1.height)}
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
