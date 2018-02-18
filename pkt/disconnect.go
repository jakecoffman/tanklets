package pkt

import (
	"github.com/jakecoffman/binser"
)

type Disconnect struct {
	ID uint16
}

func (d Disconnect) MarshalBinary() ([]byte, error) {
	return d.Serialize(nil)
}

func (d *Disconnect) UnmarshalBinary(b []byte) error {
	_, err := d.Serialize(b)
	return err
}

func (d *Disconnect) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var dc uint8 = PacketDisconnect
	stream.Uint8(&dc)
	stream.Uint16(&d.ID)
	return stream.Bytes()
}
