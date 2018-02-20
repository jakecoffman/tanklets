package server

import (
	"fmt"
	"math"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
)

func BulletPreSolve(arb *cp.Arbiter, _ *cp.Space, data interface{}) bool {
	// since bullets don't push around things, this is good to do right away
	// TODO: power-ups that make bullets non-lethal would be cool
	arb.Ignore()

	a, b := arb.Shapes()
	bullet := a.UserData.(*tanklets.Bullet)

	switch b.UserData.(type) {
	case *tanklets.Tank:
		tank := b.UserData.(*tanklets.Tank)

		// prevent bullets that were just fired from destroying the tank that fired it
		if bullet.PlayerID == tank.ID {
			if bullet.TimeAlive == 0 && bullet.Bounce == 0 {
				// this means the turret is too short, just let it go
				return true
			}
			if bullet.TimeAlive < .02 {
				// prevent player from shooting into a wall and killing themselves at short range
				// TODO: prevent with a raycast instead
				bullet.Destroy(false)
				bullet.Bounce = 100
				Players.SendAll(bullet.Location())
				return false
			}
		}

		if !tank.Destroyed {
			tank.Destroyed = true
			fmt.Println("Tank", tank.ID, "destroyed by Tank", bullet.PlayerID, "bullet", bullet.ID)
			Players.SendAll(pkt.Damage{tank.ID, bullet.PlayerID})

			// check for end game scenarios
			var tanksAlive []*tanklets.Tank
			game := data.(*Game)
			for _, t := range game.Tanks {
				if !t.Destroyed {
					tanksAlive = append(tanksAlive, t)
				}
			}
			switch {
			case len(tanksAlive) == 1:
				game.EndTime = time.Now()
				game.State = tanklets.StateWinCountdown
				Players.SendAll(pkt.State{State: tanklets.StateWinCountdown, ID: tanksAlive[0].ID})
			case len(tanksAlive) == 0:
				game.EndTime = time.Now()
				game.State = tanklets.StateFailCountdown
				Players.SendAll(pkt.State{State: tanklets.StateFailCountdown})
			}
		}

		bullet.Destroy(false)
		bullet.Bounce = 100
		Players.SendAll(bullet.Location())
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
