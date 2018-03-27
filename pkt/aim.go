package pkt

import "github.com/jakecoffman/binser"

type Aim struct {
	TurretAngle float32
}

func (a Aim) MarshalBinary() ([]byte, error) {
	return a.Serialize(nil)
}

func (a *Aim) UnmarshalBinary(b []byte) error {
	_, err := a.Serialize(b)
	return err
}

func (a *Aim) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var t uint8 = PacketAim
	stream.Uint8(&t)
	stream.Float32(&a.TurretAngle)
	return stream.Bytes()
}
