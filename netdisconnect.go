package tanklets

import (
	"fmt"
	"log"
	"net"
	"github.com/jakecoffman/binserializer"
)

type Disconnect struct {
	ID PlayerID
}

func (d Disconnect) Handle(addr *net.UDPAddr) {
	if IsServer {
		fmt.Println("SERVER: Player", d.ID, "has disconnceted")

		playerID := Lookup[addr.String()]
		player := Players.Get(playerID)
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return
		}

		Players.Delete(playerID)
		delete(Lookup, addr.String())
		Tanks[playerID].Destroyed = true

		// tell others they left & destroyed
		Players.SendAll(Disconnect{ID: playerID}, Damage{ID: playerID})
	} else {
		fmt.Println("Client", Me, "-- Player", d.ID, "Has disonnceted")
	}
}

func (d Disconnect) MarshalBinary() ([]byte, error) {
	buf := binserializer.NewBuffer(3)
	buf.WriteByte(DISCONNECT)
	buf.WriteUint16(uint16(d.ID))
	return buf.Bytes()
}

func (d *Disconnect) UnmarshalBinary(b []byte) error {
	buf := binserializer.NewBufferFromBytes(b)
	d.ID = PlayerID(buf.GetUint16())
	return buf.Error()
}
