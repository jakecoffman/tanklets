package client

import (
	"math"

	"github.com/jakecoffman/tanklets"
)

func FixedUpdate(tank *tanklets.Tank) {
	turretAngle := tank.Turret.Angle()

	tank.FixedUpdate()

	angle := float64(tank.Aim)
	diff := turretAngle - angle
	if math.Abs(diff) > 3 {
		tank.Turret.SetAngle(angle)
	} else {
		tank.Turret.SetAngle(tank.Turret.Body.Angle() - diff * .1)
	}

	tank.LastMove = tank.NextMove
}
