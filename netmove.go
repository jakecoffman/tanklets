package tanklets

import (
	"log"
	"net"
	"github.com/jakecoffman/binser"
)

// Sent to server only: Move relays inputs related to movement
type Move struct {
	Turn, Throttle int8
	TurretX, TurretY float64
}

func (m *Move) Handle(addr *net.UDPAddr) {
	tank := Tanks[Lookup[addr.String()]]
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
	return m.Serialize(nil)
}

func (m *Move) UnmarshalBinary(b []byte) error {
	_, err := m.Serialize(b)
	return err
}

func (m *Move) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var t uint8 = MOVE
	stream.Uint8(&t)
	stream.Int8(&m.Turn)
	stream.Int8(&m.Throttle)
	stream.Float64(&m.TurretX)
	stream.Float64(&m.TurretY)
	return stream.Bytes()
}
