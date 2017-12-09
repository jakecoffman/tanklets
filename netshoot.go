package tanklets

import (
	"log"
	"net"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binserializer"
)

// 54 bytes
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
		player := Players.Get(id)
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
		Players.SendAll(shot)
	} else {
		firedBy := Tanks[s.PlayerID]
		bullet := Bullets[s.BulletID]
		if bullet == nil {
			bullet = NewBullet(firedBy, s.BulletID)
			Bullets[s.BulletID] = bullet
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
	if IsServer {
		buf := binserializer.NewBuffer(55)
		buf.WriteByte(SHOOT)
		buf.WriteUint16(uint16(s.PlayerID))
		buf.WriteUint64(uint64(s.BulletID))
		buf.WriteInt16(s.Bounce)
		buf.WriteFloat64(s.X)
		buf.WriteFloat64(s.Y)
		buf.WriteFloat64(s.Vx)
		buf.WriteFloat64(s.Vy)
		buf.WriteFloat64(s.Angle)
		return buf.Bytes()
	} else {
		buf := binserializer.NewBuffer(3)
		buf.WriteByte(SHOOT)
		buf.WriteUint16(uint16(s.PlayerID))
		return buf.Bytes()
	}
}

func (s *Shoot) UnmarshalBinary(b []byte) error {
	buf := binserializer.NewBufferFromBytes(b)
	_ = buf.GetByte()
	if IsServer {
		s.PlayerID = PlayerID(buf.GetUint16())
	} else {
		s.PlayerID = PlayerID(buf.GetUint16())
		s.BulletID = BulletID(buf.GetUint64())
		s.Bounce = buf.GetInt16()
		s.X = buf.GetFloat64()
		s.Y = buf.GetFloat64()
		s.Vx = buf.GetFloat64()
		s.Vy = buf.GetFloat64()
		s.Angle = buf.GetFloat64()
	}
	return buf.Error()
}
