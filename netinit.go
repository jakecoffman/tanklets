package tanklets

import (
	"fmt"
	"net"

	"github.com/jakecoffman/binser"
)

type Init struct {
	ID    PlayerID
}

func (j *Init) Handle(addr *net.UDPAddr, game *Game) {
	if IsServer {
		id, ok := Lookup[addr.String()]
		if ok {
			j.ID = id
			fmt.Println("Player", id, "reconnected", addr)
		} else {
			id = PlayerID(game.playerIdCursor.Next())
			j.ID = id
			Lookup[addr.String()] = id
			Players.Put(id, addr)
			fmt.Println("Player", id, "connected", addr)
		}

		ServerSend(j, addr)
	} else {
		fmt.Println("I am connected!")
		Me = j.ID
		ClientIsConnected = true
		ClientIsConnecting = false
	}
}

func (j Init) MarshalBinary() ([]byte, error) {
	return j.Serialize(nil)
}

func (j *Init) UnmarshalBinary(b []byte) error {
	_, err := j.Serialize(b)
	return err
}

func (j *Init) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = INIT
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&j.ID))
	return stream.Bytes()
}
