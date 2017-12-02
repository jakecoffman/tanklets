package client

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/tanklets"
)

func DrawTank(tank *tanklets.Tank) {
	pos := tank.Position()
	x, y := float32(pos.X), float32(pos.Y)

	color := tank.Color
	if tank.Destroyed {
		color = mgl32.Vec3{0.2, 0.2, 0.2}
	}

	Renderer.DrawSprite(tankTexture, mgl32.Vec2{x, y}, mgl32.Vec2{tanklets.TankWidth, tanklets.TankHeight}, tank.Angle(), color)
	Renderer.DrawSprite(turretTexture, mgl32.Vec2{x, y}, mgl32.Vec2{32, 32}, tank.Turret.Angle(), color)
}

func DrawBullet(bullet *tanklets.Bullet) {
	pos := bullet.Body.Position()
	x, y := float32(pos.X), float32(pos.Y)
	color := tanklets.Tanks[bullet.PlayerID].Color
	Renderer.DrawSprite(bulletTexture, mgl32.Vec2{x, y}, bullet.Size(), bullet.Body.Angle(), color)
}
