package tanklets

import (
	"bytes"
	"net"
	"log"
	"github.com/jakecoffman/cp"
)

// message sent to clients: update location information
type Location struct {
	ID                     PlayerID
	X, Y                   float64
	Vx, Vy                 float64
	Angle, AngularVelocity float64

	Turret float64
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
	// ignore if the change is insignificant
	if player.Body.Position().Sub(cp.Vector{l.X, l.Y}).LengthSq() > 4 {
		player.Body.SetPosition(cp.Vector{l.X, l.Y})
	}
	player.Body.SetVelocity(l.Vx, l.Vy)
	player.ControlBody.SetVelocity(l.Vx, l.Vy)
	player.Body.SetAngle(l.Angle)
	player.ControlBody.SetAngle(l.Angle)
	player.Body.SetAngularVelocity(l.AngularVelocity)
	player.ControlBody.SetAngularVelocity(l.AngularVelocity)
	player.Turret.Body.SetAngle(l.Turret)
}

func (l *Location) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{LOCATION, byte(l.ID)})
	fields := []interface{}{&l.X, &l.Y, &l.Vx, &l.Vy, &l.Angle, &l.AngularVelocity, &l.Turret}
	return Marshal(fields, buf)
}

func (l *Location) UnmarshalBinary(b []byte) error {
	l.ID = PlayerID(b[1])
	reader := bytes.NewReader(b[2:])
	fields := []interface{}{&l.X, &l.Y, &l.Vx, &l.Vy, &l.Angle, &l.AngularVelocity, &l.Turret}
	return Unmarshal(fields, reader)
}
