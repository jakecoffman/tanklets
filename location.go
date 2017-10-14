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

func (l *Location) Handle(addr *net.UDPAddr) error {
	if IsServer {
		log.Println("I shouldn't have gotten this")
		return nil
	}

	player := Tanks[l.ID]
	if player == nil {
		log.Println("Player with ID", l.ID, "not found")
		return nil
	}
	// TODO: check if the change is insignificant and ignore it if that's the case
	player.Body.SetPosition(cp.Vector{l.X, l.Y})
	player.Body.SetVelocity(l.Vx, l.Vy)
	player.Body.SetAngle(l.Angle)
	player.Body.SetAngularVelocity(l.AngularVelocity)
	player.Turret.Body.SetAngle(l.Turret)
	return nil
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
