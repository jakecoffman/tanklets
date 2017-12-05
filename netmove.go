package tanklets

import (
	"bytes"
	"log"
	"net"
)

// Sent to server only: Move relays inputs related to movement
type Move struct {
	Turn, Throttle int8
	TurretX, TurretY float64
}

func (m *Move) Handle(addr *net.UDPAddr) {
	var tank *Tank = Tanks[Lookup[addr.String()]]
	if tank == nil {
		log.Println("Player not found", addr.String(), Lookup[addr.String()])
		return
	}
	if tank.Destroyed {
		return
	}

	tank.NextMove.Turn = m.Turn
	tank.NextMove.Throttle = m.Throttle
	tank.NextMove.TurretX = m.TurretX
	tank.NextMove.TurretY = m.TurretY
}

func (m Move) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{MOVE})
	fields := []interface{}{&m.Turn, &m.Throttle, &m.TurretX, &m.TurretY}
	return Marshal(fields, buf)
}

func (m *Move) UnmarshalBinary(b []byte) error {
	reader := bytes.NewReader(b[1:])
	fields := []interface{}{&m.Turn, &m.Throttle, &m.TurretX, &m.TurretY}
	return Unmarshal(fields, reader)
}
