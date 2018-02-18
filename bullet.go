package tanklets

import (
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"fmt"
)

const (
	BulletSpeed = 150
	BulletTTL   = 10
)

type BulletID uint64

type Bullet struct {
	ID       BulletID
	PlayerID PlayerID
	Body     *cp.Body
	Shape    *cp.Shape
	Bounce   int

	timeAlive float64

	game *Game // back reference only for removing bullets from the list... can this be done elsewhere?
}

func (g *Game) NewBullet(firedBy *Tank, id BulletID) *Bullet {
	bullet := &Bullet{
		ID:       id,
		PlayerID: firedBy.ID,
		game: g,
	}
	bullet.Body = g.Space.AddBody(cp.NewKinematicBody())
	bullet.Shape = g.Space.AddShape(cp.NewSegment(bullet.Body, cp.Vector{-2, 0}, cp.Vector{2, 0}, 3))
	//bullet.Shape.SetSensor(true)
	bullet.Shape.SetCollisionType(CollisionTypeBullet)
	bullet.Shape.SetFilter(PlayerFilter)
	bullet.Shape.UserData = bullet

	g.Bullets[bullet.ID] = bullet

	return bullet
}

func (bullet *Bullet) Update(dt float64) {
	bullet.timeAlive += dt
	if bullet.timeAlive > BulletTTL {
		// don't call Destroy because it fires after the next step
		delete(bullet.game.Bullets, bullet.ID)
		bullet.Shape.UserData = nil
		space := bullet.Shape.Space()
		space.RemoveShape(bullet.Shape)
		space.RemoveBody(bullet.Body)
		bullet.Shape = nil
		bullet.Body = nil
	}
}

func (bullet *Bullet) Size() mgl32.Vec2 {
	return mgl32.Vec2{10, 10}
}

func (bullet *Bullet) Destroy(now bool) {
	delete(bullet.game.Bullets, bullet.ID)

	if bullet.Shape == nil {
		log.Println("Shape was removed multiple times")
		return
	}

	space := bullet.Shape.Space()

	if now {
		bullet.Shape.UserData = nil
		space.RemoveShape(bullet.Shape)
		space.RemoveBody(bullet.Body)
		bullet.Shape = nil
		bullet.Body = nil
		return
	}

	// additions and removals can't be done in a normal callback.
	// Schedule a post step callback to do it.
	space.AddPostStepCallback(func(s *cp.Space, a interface{}, b interface{}) {
		if bullet.Shape == nil {
			// this fixes a crash when a tank is touching another and shoots it
			return
		}
		bullet.Shape.UserData = nil
		s.RemoveShape(bullet.Shape)
		s.RemoveBody(bullet.Body)
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

		if !tank.Destroyed {
			tank.Destroyed = true
			fmt.Println("Tank", tank.ID, "destroyed by Tank", bullet.PlayerID, "bullet", bullet.ID)
			Players.SendAll(Damage{tank.ID, bullet.PlayerID})
		}

		bullet.Destroy(false)

		bullet.Bounce = 100
		shot := bullet.Location()
		Players.SendAll(shot)
	case *Bullet:
		bullet2 := b.UserData.(*Bullet)

		bullet.Destroy(false)
		bullet2.Destroy(false)

		bullet.Bounce = 100
		bullet2.Bounce = 100

		shot1 := bullet.Location()
		shot2 := bullet2.Location()
		Players.SendAll(shot1, shot2)
	default:
		// This will bounce over anything that isn't a tank or bullet, probably check for wall here?

		bullet.Bounce++

		if bullet.Bounce > 1 {
			bullet.Destroy(false)
		} else {
			// bounce
			d := bullet.Body.Velocity()
			normal := arb.Normal()
			reflection := d.Sub(normal.Mult(2 * d.Dot(normal)))
			bullet.Body.SetVelocityVector(reflection)
			bullet.Body.SetAngle(math.Atan2(reflection.Y, reflection.X))
		}

		shot := bullet.Location()
		Players.SendAll(shot)
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
