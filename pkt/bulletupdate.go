package pkt

import (
	"github.com/jakecoffman/binser"
)

type BulletUpdate struct {
	PlayerID PlayerID
	BulletID BulletID
	Bounce   int16

	X, Y                   float64
	Vx, Vy                 float64
	Angle, AngularVelocity float64
}

func (s BulletUpdate) MarshalBinary() ([]byte, error) {
	return s.Serialize(nil)
}

func (s *BulletUpdate) UnmarshalBinary(b []byte) error {
	_, err := s.Serialize(b)
	return err
}

func (s *BulletUpdate) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var t uint8 = PacketBulletUpdate
	stream.Uint8(&t)
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
