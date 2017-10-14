package tanklets

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"time"
)

const (
	bulletSpeed = 150

	bulletTTL = 10 * time.Second
)

var Bullets = []*Bullet{}

type Bullet struct {
	Body  *cp.Body
	Shape *cp.Shape

	firedAt time.Time

	Color mgl32.Vec3
}

func NewBullet(color mgl32.Vec3) *Bullet {
	bullet := &Bullet{}
	bullet.Body = cp.NewKinematicBody()
	bullet.Shape = cp.NewCircle(bullet.Body, 5, cp.Vector{})
	bullet.Color = color
	bullet.firedAt = time.Now()

	Bullets = append(Bullets, bullet)

	return bullet
}

func (bullet *Bullet) Update() {
}

func (bullet *Bullet) Size() mgl32.Vec2 {
	return mgl32.Vec2{10, 10}
}
