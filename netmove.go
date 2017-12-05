package tanklets

import (
	"bytes"
	"log"
	"net"

	"github.com/jakecoffman/cp"
)

// Sent to server only: Move relays inputs related to movement
type Move struct {
	PlayerID               PlayerID // 2 bytes
	Turn, Throttle, Turret float64  // 24 bytes
}

const maxTurn = 0.1

func (m *Move) Handle(addr *net.UDPAddr) {
	var tank *Tank = Tanks[Lookup[addr.String()]]
	if tank == nil {
		log.Println("Player not found", addr.String(), Lookup[addr.String()])
		return
	}

	if m.Turn > maxTurn {
		log.Println("Player tried to turn too fast: cheating?", m.Turn)
		return
	}

	ApplyMove(tank, m)
}

func ApplyMove(tank *Tank, m *Move) {
	tank.ControlBody.SetAngle(tank.Body.Angle() + m.Turn)
	// by applying to the body too, it will allow getting unstuck from corners
	tank.Body.SetAngle(tank.Body.Angle() + m.Turn)

	if m.Throttle != 0 {
		tank.ControlBody.SetVelocityVector(tank.Body.Rotation().Rotate(cp.Vector{Y: m.Throttle * MaxSpeed}))
	} else {
		tank.ControlBody.SetVelocity(0, 0)
	}

	tank.Turret.SetAngle(tank.Turret.Angle() - m.Turret)
	tank.Turret.SetPosition(tank.Body.Position())
}

func (m Move) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{MOVE})
	if IsServer {
		fields := []interface{}{&m.PlayerID, &m.Turn, &m.Throttle, &m.Turret}
		return Marshal(fields, buf)
	} else {
		fields := []interface{}{&m.Turn, &m.Throttle, &m.Turret}
		return Marshal(fields, buf)
	}
}

func (m *Move) UnmarshalBinary(b []byte) error {
	reader := bytes.NewReader(b[1:])
	if IsServer {
		fields := []interface{}{&m.Turn, &m.Throttle, &m.Turret}
		return Unmarshal(fields, reader)
	} else {
		fields := []interface{}{&m.PlayerID, &m.Turn, &m.Throttle, &m.Turret}
		return Unmarshal(fields, reader)
	}
}
