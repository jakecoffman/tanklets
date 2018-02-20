package pkt

import (
	"github.com/jakecoffman/binser"
)

type State struct {
	State uint8
	ID    PlayerID // player that won
}

func (j State) MarshalBinary() ([]byte, error) {
	return j.Serialize(nil)
}

func (j *State) UnmarshalBinary(b []byte) error {
	_, err := j.Serialize(b)
	return err
}

func (j *State) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = PacketState
	stream.Uint8(&m)
	stream.Uint8(&j.State)
	stream.Uint16((*uint16)(&j.ID))
	return stream.Bytes()
}
