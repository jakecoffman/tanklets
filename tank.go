package tanklets

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

const (
	turretw = 4
	turreth = 15

	turnSpeed = .05
	maxSpeed  = 60

	shotCooldown = 250 * time.Millisecond
)

type Tank struct {
	Turret
	*cp.Body
	*cp.Shape

	tankTexture, turretTexture, bulletTexture *Texture2D

	width, height float64
	color         mgl32.Vec3

	ControlBody *cp.Body
	LastShot    time.Time
}

type Turret struct {
	*cp.Body
	*cp.Shape
}

func NewTank(space *cp.Space, tankTex, turretTex, bulletTex *Texture2D, w, h int) Tank {
	width := float64(w)
	height := float64(h)
	tank := Tank{
		width:         width,
		height:        height,
		tankTexture:   tankTex,
		turretTexture: turretTex,
		bulletTexture: bulletTex,
	}
	tank.ControlBody = space.AddBody(cp.NewKinematicBody())
	tank.Body = space.AddBody(cp.NewBody(1, cp.MomentForBox(1, width, height)))
	tankShape := space.AddShape(cp.NewBox(tank.Body, width, height, 2))
	tankShape.SetElasticity(0)
	tankShape.SetFriction(0)
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

func (tank *Tank) Update() {
	// update body

	// update turret
	tank.Turret.SetPosition(tank.Body.Position())
	mouseDelta := Mouse.Sub(tank.Turret.Body.Position())
	turn := tank.Turret.Rotation().Unrotate(mouseDelta).ToAngle()
	tank.Turret.SetAngle(tank.Turret.Angle() - turn)
}

func (tank *Tank) Draw(renderer *SpriteRenderer) {
	pos := tank.Position()
	x, y := float32(pos.X), float32(pos.Y)
	renderer.DrawSprite(tank.tankTexture, mgl32.Vec2{x, y}, tank.Size(), tank1.Angle(), tank.color)

	renderer.DrawSprite(tank.turretTexture, mgl32.Vec2{x, y}, tank.Turret.Size(), tank1.Turret.Angle(), tank.color)
}

func (tank *Tank) Size() mgl32.Vec2 {
	return mgl32.Vec2{float32(tank1.width), float32(tank1.height)}
}

func (turret *Turret) Size() mgl32.Vec2 {
	return mgl32.Vec2{float32(tank1.height), float32(tank1.height)}
}

func (tank *Tank) Shoot(space *cp.Space) {
	bullet := NewBullet(tank.color, tank.bulletTexture)

	pos := cp.Vector{X: float64(tank.Turret.Size().Y()/2)}
	pos = pos.Rotate(tank.Turret.Rotation())
	bullet.Body.SetPosition(pos.Add(tank.Turret.Position()))
	bullet.Body.SetAngle(tank.Turret.Angle())
	bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{bulletSpeed, 0}))
	bullet.Shape.SetFilter(cp.NewShapeFilter(1, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	space.AddBody(bullet.Body)
	space.AddShape(bullet.Shape)
}
