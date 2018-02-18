package tanklets

import (
	"github.com/jakecoffman/binser"
)

type Initial struct {
	ID    PlayerID
}

func (j Initial) MarshalBinary() ([]byte, error) {
	return j.Serialize(nil)
}

func (j *Initial) UnmarshalBinary(b []byte) error {
	_, err := j.Serialize(b)
	return err
}

func (j *Initial) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = PacketInit
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&j.ID))
	return stream.Bytes()
}
