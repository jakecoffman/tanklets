package tanklets

import (
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

const (
	bulletSpeed = 150
	bulletTTL   = 10
)

type BulletID uint64

var Bullets = map[BulletID]*Bullet{}
var bulletCurId BulletID

type Bullet struct {
	ID       BulletID
	PlayerID PlayerID
	Body     *cp.Body
	Shape    *cp.Shape
	Bounce   int

	timeAlive float64
}

func NewBullet(firedBy *Tank, id BulletID) *Bullet {
	bullet := &Bullet{
		ID:       id,
		PlayerID: firedBy.ID,
	}
	bullet.Body = Space.AddBody(cp.NewKinematicBody())
	bullet.Shape = Space.AddShape(cp.NewCircle(bullet.Body, 5, cp.Vector{}))
	//bullet.Shape.SetSensor(true)
	bullet.Shape.SetCollisionType(COLLISION_TYPE_BULLET)
	bullet.Shape.SetFilter(PlayerFilter)
	bullet.Shape.UserData = bullet

	Bullets[bullet.ID] = bullet

	return bullet
}

func (bullet *Bullet) Update(dt float64) {
	bullet.timeAlive += dt
	if bullet.timeAlive > bulletTTL {
		// don't call Destroy because it fires after the next step
		delete(Bullets, bullet.ID)
		bullet.Shape.UserData = nil
		Space.RemoveShape(bullet.Shape)
		Space.RemoveBody(bullet.Body)
		bullet.Shape = nil
		bullet.Body = nil
	}
}

func (bullet *Bullet) Size() mgl32.Vec2 {
	return mgl32.Vec2{10, 10}
}

func (bullet *Bullet) Destroy() {
	delete(Bullets, bullet.ID)

	if bullet.Shape == nil {
		log.Println("Shape was removed multiple times")
		return
	}

	// additions and removals can't be done in a normal callback.
	// Schedule a post step callback to do it.
	Space.AddPostStepCallback(func(space *cp.Space, a interface{}, b interface{}) {
		bullet.Shape.UserData = nil
		space.RemoveShape(bullet.Shape)
		space.RemoveBody(bullet.Body)
		bullet.Shape = nil
		bullet.Body = nil
	}, nil, nil)
}

func BulletPreSolve(arb *cp.Arbiter, _ *cp.Space, _ interface{}) bool {
	// since bullets don't push around things, this is good to do right away
	// TODO: power-ups that make bullets non-lethal would be cool
	arb.Ignore()

	// clients don't decide this stuff, so just ignore
	// TODO: Don't set this custom callback for clients at all
	if !IsServer {
		return false
	}

	a, b := arb.Shapes()
	bullet := a.UserData.(*Bullet)

	switch b.UserData.(type) {
	case *Tank:
		tank := b.UserData.(*Tank)

		// Before first bounce, tank can't hit itself
		if bullet.Bounce < 1 && bullet.PlayerID == tank.ID {
			return false
		}

		tank.Damage(bullet)
		bullet.Destroy()

		bullet.Bounce = 100
		shot := bullet.Location()
		for _, p := range Players {
			Send(shot, p)
		}
	case *Bullet:
		bullet2 := b.UserData.(*Bullet)

		bullet.Destroy()
		bullet2.Destroy()

		bullet.Bounce = 100
		bullet2.Bounce = 100

		shot1 := bullet.Location()
		shot2 := bullet2.Location()
		for _, p := range Players {
			Send(shot1, p)
			Send(shot2, p)
		}
	default:
		// TODO: This will bounce over anything that isn't a tank or bullet
		bullet.Bounce++

		if bullet.Bounce > 1 {
			bullet.Destroy()
			return false
		}

		d := bullet.Body.Velocity()
		normal := arb.Normal()
		reflection := d.Sub(normal.Mult(2 * d.Dot(normal)))
		bullet.Body.SetVelocityVector(reflection)
		bullet.Body.SetAngle(math.Atan2(reflection.Y, reflection.X))

		// do I need to apply an impulse here to get it out of whatever it hit?

		shot := bullet.Location()
		for _, p := range Players {
			Send(shot, p)
		}
	}
	return false
}

func (bullet *Bullet) Location() *Shoot {
	return &Shoot{
		BulletID:        bullet.ID,
		PlayerID:        bullet.PlayerID,
		Bounce:          int16(bullet.Bounce),
		X:               bullet.Body.Position().X,
		Y:               bullet.Body.Position().Y,
		Angle:           bullet.Body.Angle(),
		AngularVelocity: bullet.Body.AngularVelocity(),
		Vx:              bullet.Body.Velocity().X,
		Vy:              bullet.Body.Velocity().Y,
	}
}
