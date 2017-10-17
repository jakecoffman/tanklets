package tanklets

import (
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
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

	Tank   *Tank
	Bounce int
}

func NewBullet(firedBy *Tank) *Bullet {
	bullet := &Bullet{}
	bullet.Body = cp.NewKinematicBody()
	bullet.Shape = bullet.Body.AddShape(cp.NewCircle(bullet.Body, 5, cp.Vector{}))
	//bullet.Shape.SetSensor(true)
	bullet.Shape.SetCollisionType(COLLISION_TYPE_BULLET)
	bullet.Shape.SetFilter(PlayerFilter)
	bullet.Shape.UserData = bullet
	bullet.Tank = firedBy
	bullet.firedAt = time.Now()

	Bullets = append(Bullets, bullet)

	return bullet
}

func (bullet *Bullet) Update() {
}

func (bullet *Bullet) Size() mgl32.Vec2 {
	return mgl32.Vec2{10, 10}
}

func (bullet *Bullet) Destroy() {
	bullet.Shape.UserData = nil
	Space.RemoveBody(bullet.Body)
	Space.RemoveShape(bullet.Shape)
	bullet.Tank = nil
}

func BulletPreSolve(arb *cp.Arbiter, _ *cp.Space, _ interface{}) bool {
	a, b := arb.Shapes()
	bullet := a.UserData.(*Bullet)

	switch b.UserData.(type) {
	case *Tank:
		tank := b.UserData.(*Tank)

		if bullet.Bounce < 1 && bullet.Tank == tank {
			return arb.Ignore()
		}

		if IsServer {
			tank.Damage(bullet)
		}
	default:
	}

	for i, b := range Bullets {
		if b == bullet {
			Bullets = append(Bullets[:i], Bullets[i+1:]...)
			// additions and removals can't be done in a normal callback.
			// Schedule a post step callback to do it.
			Space.AddPostStepCallback(func (*cp.Space, interface{}, interface{}) {
				bullet.Destroy()
			}, nil, nil)
			break
		}
	}
	return false
}
