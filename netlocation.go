package tanklets

import (
	"github.com/jakecoffman/binser"
)

// message sent to clients: update location information
type Location struct {
	ID                     PlayerID
	X, Y                   float32
	Vx, Vy                 float32
	Angle, AngularVelocity float32

	Turret float32
}

func (l Location) MarshalBinary() ([]byte, error) {
	return l.Serialize(nil)
}

func (l *Location) UnmarshalBinary(b []byte) error {
	_, err := l.Serialize(b)
	return err
}

func (l *Location) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = PacketLocation
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&l.ID))
	stream.Float32(&l.X)
	stream.Float32(&l.Y)
	stream.Float32(&l.Vx)
	stream.Float32(&l.Vy)
	stream.Float32(&l.Angle)
	stream.Float32(&l.AngularVelocity)
	stream.Float32(&l.Turret)
	return stream.Bytes()
}
