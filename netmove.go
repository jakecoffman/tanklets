package tanklets

import (
	"log"
	"net"
	"github.com/jakecoffman/binser"
)

// Sent to server only: Move relays inputs related to movement
type Move struct {
	Turn, Throttle int8
	TurretAngle float64
}

func (m *Move) Handle(addr *net.UDPAddr, game *Game) {
	if game.State != GameStatePlaying {
		return
	}

	tank := game.Tanks[Lookup[addr.String()]]
	if tank == nil {
		log.Println("Player not found", addr.String(), Lookup[addr.String()])
		return
	}
	if tank.Destroyed {
		return
	}

	tank.NextMove.Turn = m.Turn
	tank.NextMove.Throttle = m.Throttle
	tank.NextMove.TurretAngle = m.TurretAngle
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
	var atRest uint8
	if !stream.IsReading() && m.Turn == 0 && m.Throttle == 0 {
		atRest = 1
	}
	stream.Uint8(&atRest)
	if atRest == 0 {
		stream.Int8(&m.Turn)
		stream.Int8(&m.Throttle)
	} else if stream.IsReading() {
		m.Turn = 0
		m.Throttle = 0
	}
	stream.Float64(&m.TurretAngle)
	return stream.Bytes()
}
