package pkt

import (
	"github.com/jakecoffman/binser"
)

type Damage struct {
	ID PlayerID
	Killer PlayerID
}

func (d Damage) MarshalBinary() ([]byte, error) {
	return d.Serialize(nil)
}

func (d *Damage) UnmarshalBinary(b []byte) error {
	_, err := d.Serialize(b)
	return err
}

func (d *Damage) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = PacketDamage
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&d.ID))
	stream.Uint16((*uint16)(&d.Killer))
	return stream.Bytes()
}
