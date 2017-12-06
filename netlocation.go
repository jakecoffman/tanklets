package tanklets

import (
	"bytes"
	"log"
	"net"

	"math"

	"github.com/jakecoffman/cp"
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
