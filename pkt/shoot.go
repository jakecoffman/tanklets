package pkt

import (
	"github.com/jakecoffman/binser"
)

type Shoot struct {
	Angle float64
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
	stream.Float64(&s.Angle)
	return stream.Bytes()
}
