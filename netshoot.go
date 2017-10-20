package tanklets

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/jakecoffman/cp"
)

type Shoot struct {
	PlayerID PlayerID
	BulletID BulletID
	Bounce   int16

	X, Y                   float64
	Vx, Vy                 float64
	Angle, AngularVelocity float64
}

func (s *Shoot) Handle(addr *net.UDPAddr) {
	if IsServer {
		id := Lookup[addr.String()]
		player := Players[id]
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return
		}
		tank := Tanks[id]

		if time.Now().Sub(tank.LastShot) < ShotCooldown {
			return
		}
		tank.LastShot = time.Now()

		bullet := NewBullet(tank, bulletCurId)
		bulletCurId++

		pos := cp.Vector{X: TankHeight / 2.0}
		pos = pos.Rotate(tank.Turret.Rotation())
		bullet.Body.SetPosition(pos.Add(tank.Turret.Position()))
		bullet.Body.SetAngle(tank.Turret.Angle())
		bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{bulletSpeed, 0}))
		//bullet.Shape.SetFilter(cp.NewShapeFilter(uint(player.ID), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

		shot := bullet.Location()
		for _, p := range Players {
			Send(shot, p)
		}
	} else {
		firedBy := Tanks[s.PlayerID]
		bullet := Bullets[s.BulletID]
		if bullet == nil {
			bullet = NewBullet(firedBy, s.BulletID)
			Bullets[s.BulletID] = bullet
		}

		bullet.Bounce = int(s.Bounce)

		if bullet.Bounce > 1 {
			bullet.Destroy()
			return
		}

		bullet.Body.SetPosition(cp.Vector{s.X, s.Y})
		bullet.Body.SetAngle(s.Angle)
		bullet.Body.SetVelocity(s.Vx, s.Vy)
	}
}

func (s Shoot) MarshalBinary() ([]byte, error) {
	if IsServer {
		buf := bytes.NewBuffer([]byte{SHOOT})
		fields := []interface{}{
			&s.PlayerID, &s.BulletID, &s.Bounce, &s.X, &s.Y, &s.Vx, &s.Vy, &s.Angle,
		}
		return Marshal(fields, buf)
	} else {
		buf := make([]byte, 3)
		buf[0] = SHOOT
		binary.BigEndian.PutUint16(buf[1:3], uint16(s.PlayerID))
		return buf, nil
	}
}

func (s *Shoot) UnmarshalBinary(buf []byte) error {
	if IsServer {
		s.PlayerID = PlayerID(binary.BigEndian.Uint16(buf[1:3]))
		return nil
	} else {
		reader := bytes.NewReader(buf[1:])
		fields := []interface{}{
			&s.PlayerID, &s.BulletID, &s.Bounce, &s.X, &s.Y, &s.Vx, &s.Vy, &s.Angle,
		}
		return Unmarshal(fields, reader)
	}
}
