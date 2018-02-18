package pkt

import (
	"github.com/jakecoffman/binser"
)

type Ready struct {}

func (j Ready) MarshalBinary() ([]byte, error) {
	return j.Serialize(nil)
}

func (j *Ready) UnmarshalBinary(b []byte) error {
	_, err := j.Serialize(b)
	return err
}

func (j *Ready) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = PacketReady
	stream.Uint8(&m)
	return stream.Bytes()
}
