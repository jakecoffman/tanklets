package pkt

import (
	"github.com/jakecoffman/binser"
)

// message sent to clients: update location information
type BoxLocation struct {
	ID                     BoxID
	X, Y                   float32
	Angle float32
}

func (l BoxLocation) MarshalBinary() ([]byte, error) {
	return l.Serialize(nil)
}

func (l *BoxLocation) UnmarshalBinary(b []byte) error {
	_, err := l.Serialize(b)
	return err
}

func (l *BoxLocation) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = PacketBoxLocation
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&l.ID))
	stream.Float32(&l.X)
	stream.Float32(&l.Y)
	stream.Float32(&l.Angle)
	return stream.Bytes()
}
