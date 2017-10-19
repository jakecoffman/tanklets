package tanklets

import (
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/jakecoffman/cp"
	"bytes"
)

type Shoot struct {
	ID PlayerID

	Position, Velocity cp.Vector
	Angle              float64
}

func (s *Shoot) Handle(addr *net.UDPAddr) error {
	if IsServer {
		var player *Tank = Tanks[Lookup[addr.String()]]
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return nil
		}

		if time.Now().Sub(player.LastShot) < ShotCooldown {
			return nil
		}
		player.LastShot = time.Now()

		bullet := NewBullet(player)

		pos := cp.Vector{X: TankHeight / 2.0}
		pos = pos.Rotate(player.Turret.Rotation())
		bullet.Body.SetPosition(pos.Add(player.Turret.Position()))
		bullet.Body.SetAngle(player.Turret.Angle())
		bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{bulletSpeed, 0}))
		//bullet.Shape.SetFilter(cp.NewShapeFilter(uint(player.ID), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

		Space.AddBody(bullet.Body)
		Space.AddShape(bullet.Shape)

		shot := Shoot{
			ID:       player.ID,
			Position: bullet.Body.Position(),
			Velocity: bullet.Body.Velocity(),
			Angle:    bullet.Body.Angle(),
		}
		for _, p := range Players {
			Send(shot, p)
		}
	} else {
		player := Tanks[s.ID]
		bullet := NewBullet(player)

		bullet.Body.SetPosition(s.Position)
		bullet.Body.SetAngle(s.Angle)
		bullet.Body.SetVelocityVector(s.Velocity)

		Space.AddBody(bullet.Body)
		Space.AddShape(bullet.Shape)
	}

	return nil
}

func (s Shoot) MarshalBinary() ([]byte, error) {
	if IsServer {
		buf := bytes.NewBuffer([]byte{SHOOT})
		fields := []interface{}{&s.ID, &s.Position.X, &s.Position.Y, &s.Velocity.X, &s.Velocity.Y, &s.Angle}
		return Marshal(fields, buf)
	} else {
		buf := make([]byte, 3)
		buf[0] = SHOOT
		binary.BigEndian.PutUint16(buf[1:3], uint16(s.ID))
		return buf, nil
	}
}

func (s *Shoot) UnmarshalBinary(buf []byte) error {
	if IsServer {
		s.ID = PlayerID(binary.BigEndian.Uint16(buf[1:3]))
		return nil
	} else {
		reader := bytes.NewReader(buf[1:])
		fields := []interface{}{&s.ID, &s.Position.X, &s.Position.Y, &s.Velocity.X, &s.Velocity.Y, &s.Angle}
		return Unmarshal(fields, reader)
	}
}
