package server

import (
	"github.com/jakecoffman/cp"
	"fmt"
	"github.com/jakecoffman/tanklets/pkt"
	"github.com/jakecoffman/tanklets"
	"math"
)

func BulletPreSolve(arb *cp.Arbiter, _ *cp.Space, _ interface{}) bool {
	// since bullets don't push around things, this is good to do right away
	// TODO: power-ups that make bullets non-lethal would be cool
	arb.Ignore()

	a, b := arb.Shapes()
	bullet := a.UserData.(*tanklets.Bullet)

	switch b.UserData.(type) {
	case *tanklets.Tank:
		tank := b.UserData.(*tanklets.Tank)

		// Before first bounce, tank can't hit itself
		if bullet.Bounce < 1 && bullet.PlayerID == tank.ID {
			return false
		}

		if !tank.Destroyed {
			tank.Destroyed = true
			fmt.Println("Tank", tank.ID, "destroyed by Tank", bullet.PlayerID, "bullet", bullet.ID)
			Players.SendAll(pkt.Damage{tank.ID, bullet.PlayerID})
		}

		bullet.Destroy(false)

		bullet.Bounce = 100
		shot := bullet.Location()
		Players.SendAll(shot)
	case *tanklets.Bullet:
		bullet2 := b.UserData.(*tanklets.Bullet)

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
