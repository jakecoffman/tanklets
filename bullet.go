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

	texture *Texture2D
	color   mgl32.Vec3
}

func NewBullet(color mgl32.Vec3, texture *Texture2D) *Bullet {
	bullet := &Bullet{}
	bullet.Body = cp.NewKinematicBody()
	bullet.Shape = cp.NewCircle(bullet.Body, 5, cp.Vector{})
	bullet.color = color
	bullet.texture = texture
	bullet.firedAt = time.Now()

	Bullets = append(Bullets, bullet)

	return bullet
}

func (bullet *Bullet) Update() {

}

func (bullet *Bullet) Draw(renderer *SpriteRenderer) {
	pos := bullet.Body.Position()
	x, y := float32(pos.X), float32(pos.Y)
	renderer.DrawSprite(bullet.texture, mgl32.Vec2{x, y}, bullet.Size(), bullet.Body.Angle(), bullet.color)
}

func (bullet *Bullet) Size() mgl32.Vec2 {
	return mgl32.Vec2{10, 10}
}
