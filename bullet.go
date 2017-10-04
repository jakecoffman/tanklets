package tanklets

import "github.com/jakecoffman/cp"

const (
	bulletSpeed = 150
)

type Bullet struct {
	Body *cp.Body
	Shape *cp.Shape
}

func NewBullet() *Bullet {
	bullet := &Bullet{}
	bullet.Body = cp.NewKinematicBody()
	bullet.Shape = cp.NewCircle(bullet.Body, 5, cp.Vector{})
	return bullet
}

func (bullet *Bullet) Update() {

}
