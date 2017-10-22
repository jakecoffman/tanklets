package tanklets

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
)

type Disconnect struct {
	ID PlayerID
}

func (d *Disconnect) Handle(addr *net.UDPAddr) {
	if IsServer {
		fmt.Println("SERVER: Player", d.ID, "has disonnceted")

		playerID := Lookup[addr.String()]
		var player *net.UDPAddr = Players[playerID]
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return
		}

		delete(Players, playerID)
		delete(Lookup, addr.String())
		Tanks[playerID].Destroyed = true

		// tell others they left & destroyed
		for _, p := range Players {
			Send(Disconnect{ID: playerID}, p)
			Send(Damage{ID: playerID}, p)
		}
	} else {
		fmt.Println("Client", Me, "-- Player", d.ID, "Has disonnceted")
	}
}

func (d Disconnect) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3)
	buf[0] = DISCONNECT
	binary.BigEndian.PutUint16(buf[1:3], uint16(d.ID))
	return buf, nil
}

func (d *Disconnect) UnmarshalBinary(buf []byte) error {
	d.ID = PlayerID(binary.BigEndian.Uint16(buf[1:3]))
	return nil
}
