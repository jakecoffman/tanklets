package tanklets

import (
	"net"

	"github.com/jakecoffman/binser"
)

type State struct {
	state uint8
}

func (j *State) Handle(addr *net.UDPAddr, game *Game) {
	game.State = int(j.state)
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
	var m uint8 = STATE
	stream.Uint8(&m)
	stream.Uint8(&j.state)
	return stream.Bytes()
}
