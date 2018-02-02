package tanklets

import (
	"log"
	"net"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binser"
)

// message sent to clients: update location information
type BoxLocation struct {
	ID                     BoxID
	X, Y                   float32
	Angle float32
}

func (l *BoxLocation) Handle(addr *net.UDPAddr, game *Game) {
	if IsServer {
		log.Println("I shouldn't have gotten this")
		return
	}

	box := game.Boxes[l.ID]
	if box == nil {
		box = game.NewBox(l.ID)
	}
	pos := box.Position()
	newPos := cp.Vector{float64(l.X), float64(l.Y)}

	diff := newPos.Sub(pos)
	distance := diff.Length()

	if distance > 4 {
		box.SetPosition(newPos)
	} else {
		box.SetPosition(pos.Add(diff.Mult(0.1)))
	}

	box.SetAngle(float64(l.Angle))
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
	var m uint8 = BOXLOCATION
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&l.ID))
	stream.Float32(&l.X)
	stream.Float32(&l.Y)
	stream.Float32(&l.Angle)
	return stream.Bytes()
}
