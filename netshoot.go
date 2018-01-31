package tanklets

import (
	"log"
	"net"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binser"
)

type Shoot struct {
	PlayerID PlayerID
	BulletID BulletID
	Bounce   int16

	X, Y                   float64
	Vx, Vy                 float64
	Angle, AngularVelocity float64
}

func (s *Shoot) Handle(addr *net.UDPAddr, game *Game) {
	if IsServer {
		id := Lookup[addr.String()]
		player := Players.Get(id)
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return
		}
		tank := game.Tanks[id]

		if time.Now().Sub(tank.LastShot) < ShotCooldown {
			return
		}
		tank.LastShot = time.Now()

		bullet := game.NewBullet(tank, BulletID(game.bullet.Next()))

		pos := cp.Vector{X: TankHeight / 2.0}
		pos = pos.Rotate(tank.Turret.Rotation())
		bullet.Body.SetPosition(pos.Add(tank.Turret.Position()))
		bullet.Body.SetAngle(tank.Turret.Angle())
		bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{bulletSpeed, 0}))
		//bullet.Shape.SetFilter(cp.NewShapeFilter(uint(player.ID), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

		shot := bullet.Location()
		Players.SendAll(shot)
	} else {
		firedBy := game.Tanks[s.PlayerID]
		bullet := game.Bullets[s.BulletID]
		if bullet == nil {
			bullet = game.NewBullet(firedBy, s.BulletID)
			game.Bullets[s.BulletID] = bullet
		}

		bullet.Bounce = int(s.Bounce)

		if bullet.Bounce > 1 {
			bullet.Destroy(true)
			return
		}

		bullet.Body.SetPosition(cp.Vector{s.X, s.Y})
		bullet.Body.SetAngle(s.Angle)
		bullet.Body.SetVelocity(s.Vx, s.Vy)
	}
}

func (s Shoot) MarshalBinary() ([]byte, error) {
	return s.Serialize(nil)
}

func (s *Shoot) UnmarshalBinary(b []byte) error {
	_, err := s.Serialize(b)
	return err
}

func (s *Shoot) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var t uint8 = SHOOT
	stream.Uint8(&t)
	if !IsServer && !stream.IsReading() || IsServer && stream.IsReading() {
		// the player sends this empty message to shoot
		return stream.Bytes()
	}
	// the server sends all players the rest of the data
	stream.Uint16((*uint16)(&s.PlayerID))
	stream.Uint64((*uint64)(&s.BulletID))
	stream.Int16(&s.Bounce)
	stream.Float64(&s.X)
	stream.Float64(&s.Y)
	stream.Float64(&s.Vx)
	stream.Float64(&s.Vy)
	stream.Float64(&s.Angle)
	return stream.Bytes()
}
