package pkt

import (
	"github.com/jakecoffman/binser"
)

// Sent to server only: Move relays inputs related to movement
type Move struct {
	Turn, Throttle int8
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
	var t uint8 = PacketMove
	stream.Uint8(&t)
	stream.Int8(&m.Turn)
	stream.Int8(&m.Throttle)
	return stream.Bytes()
}
