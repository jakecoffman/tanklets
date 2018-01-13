package tanklets

import (
	"log"
	"net"
	"math"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binser"
)

// message sent to clients: update location information (58 bytes)
type Location struct {
	ID                     PlayerID // 2 bytes
	X, Y                   float64  // 16 bytes
	Vx, Vy                 float64  // 16
	Angle, AngularVelocity float64  // 16

	Turret float64 // 8
}

func (l *Location) Handle(addr *net.UDPAddr) {
	if IsServer {
		log.Println("I shouldn't have gotten this")
		return
	}

	player := Tanks[l.ID]
	if player == nil {
		log.Println("Client", Me, "-- Player with ID", l.ID, "not found")
		return
	}
	pos := player.Body.Position()
	newPos := cp.Vector{l.X, l.Y}

	if pos.Distance(newPos) > 1 {
		player.Body.SetPosition(newPos)
	} else {
		player.Body.SetPosition(pos.Lerp(newPos, 0.5))
	}

	angle := player.ControlBody.Angle()
	if math.Abs(angle-l.Angle) > 5 {
		player.ControlBody.SetAngle(l.Angle)
	} else {
		player.ControlBody.SetAngle(cp.Lerp(l.Angle, angle, 0.5))
	}

	player.Turret.SetPosition(player.Body.Position())
	player.ControlBody.SetVelocity(l.Vx, l.Vy)
	player.ControlBody.SetAngularVelocity(l.AngularVelocity)
	player.Turret.Body.SetAngle(l.Turret)
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
	var m uint8 = LOCATION
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&l.ID))
	stream.Float64(&l.X)
	stream.Float64(&l.Y)
	stream.Float64(&l.Vx)
	stream.Float64(&l.Vy)
	stream.Float64(&l.Angle)
	stream.Float64(&l.AngularVelocity)
	stream.Float64(&l.Turret)
	return stream.Bytes()
}
