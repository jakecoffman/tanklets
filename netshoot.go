package tanklets

import (
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

func (s Shoot) MarshalBinary() ([]byte, error) {
	return s.Serialize(nil)
}

func (s *Shoot) UnmarshalBinary(b []byte) error {
	_, err := s.Serialize(b)
	return err
}

func (s *Shoot) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var t uint8 = PacketShoot
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
