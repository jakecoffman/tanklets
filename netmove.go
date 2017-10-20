package tanklets

import (
	"bytes"
	"log"
	"net"

	"github.com/jakecoffman/cp"
)

// Sent to server only: Move relays inputs related to movement
type Move struct {
	Turn, Throttle, Turret float64
}

const maxTurn = 0.1

func (m *Move) Handle(addr *net.UDPAddr) {
	if !IsServer {
		log.Println("I shouldn't have gotten this")
		return
	}

	var player *Tank = Tanks[Lookup[addr.String()]]
	if player == nil {
		log.Println("Player not found", addr.String(), Lookup[addr.String()])
		return
	}

	if m.Turn > maxTurn {
		log.Println("Player tried to turn too fast: cheating?", m.Turn)
		return
	}

	player.ControlBody.SetAngle(player.Body.Angle() + m.Turn)
	// by applying to the body too, it will allow getting unstuck from corners
	player.Body.SetAngle(player.Body.Angle() + m.Turn)

	if m.Throttle != 0 {
		player.ControlBody.SetVelocityVector(player.Body.Rotation().Rotate(cp.Vector{Y: m.Throttle * MaxSpeed}))
	} else {
		player.ControlBody.SetVelocity(0, 0)
	}

	player.Turret.SetAngle(player.Turret.Angle() - m.Turret)
}

func (m Move) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{MOVE})
	fields := []interface{}{&m.Turn, &m.Throttle, &m.Turret}
	return Marshal(fields, buf)
}

func (m *Move) UnmarshalBinary(b []byte) error {
	reader := bytes.NewReader(b[1:])
	fields := []interface{}{&m.Turn, &m.Throttle, &m.Turret}
	return Unmarshal(fields, reader)
}
