package tanklets

import (
	"log"
	"net"
	"github.com/jakecoffman/binserializer"
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
	buf := binserializer.NewBuffer(19)
	buf.WriteByte(MOVE)

	buf.WriteInt8(m.Turn)
	buf.WriteInt8(m.Throttle)

	buf.WriteFloat64(m.TurretX)
	buf.WriteFloat64(m.TurretY)
	return buf.Bytes()
}

func (m *Move) UnmarshalBinary(b []byte) error {
	buf := binserializer.NewBufferFromBytes(b)
	_ = buf.GetByte()

	m.Turn = buf.GetInt8()
	m.Throttle = buf.GetInt8()

	m.TurretX = buf.GetFloat64()
	m.TurretY = buf.GetFloat64()
	return buf.Error()
}
