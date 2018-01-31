package tanklets

import (
	"log"
	"net"

	"github.com/jakecoffman/cp"
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

func (l *Location) Handle(addr *net.UDPAddr, game *Game) {
	if IsServer {
		log.Println("I shouldn't have gotten this")
		return
	}

	player := game.Tanks[l.ID]
	if player == nil {
		log.Println("Client", Me, "-- Player with ID", l.ID, "not found")
		return
	}
	pos := player.Position()
	newPos := cp.Vector{float64(l.X), float64(l.Y)}

	diff := newPos.Sub(pos)
	distance := diff.Length()

	// https://gafferongames.com/post/networked_physics_2004/
	if distance > 4 {
		player.SetPosition(newPos)
	} else {
		player.SetPosition(pos.Add(diff.Mult(0.1)))
	}
	player.Turret.SetPosition(player.Body.Position())

	player.SetAngle(float64(l.Angle))
	player.ControlBody.SetAngle(player.Angle())

	player.SetVelocity(float64(l.Vx), float64(l.Vy))
	player.ControlBody.SetVelocityVector(player.Velocity())
	player.SetAngularVelocity(float64(l.AngularVelocity))
	player.ControlBody.SetAngularVelocity(player.AngularVelocity())
	player.Turret.Body.SetAngle(float64(l.Turret))
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
	stream.Float32(&l.X)
	stream.Float32(&l.Y)
	stream.Float32(&l.Vx)
	stream.Float32(&l.Vy)
	stream.Float32(&l.Angle)
	stream.Float32(&l.AngularVelocity)
	stream.Float32(&l.Turret)
	return stream.Bytes()
}
