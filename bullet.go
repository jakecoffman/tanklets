package tanklets

import (
	"log"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets/pkt"
)

const (
	BulletSpeed = 150
	BulletTTL   = 10
)

type Bullet struct {
	ID       BulletID
	PlayerID PlayerID
	Body     *cp.Body
	Shape    *cp.Shape
	Bounce   int

	TimeAlive float64

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
	bullet.TimeAlive += dt
	if bullet.TimeAlive > BulletTTL {
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

func (bullet *Bullet) Location() *pkt.BulletUpdate {
	return &pkt.BulletUpdate{
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
