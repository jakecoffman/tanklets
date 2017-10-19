package tanklets

import (
	"encoding/binary"
	"log"
	"net"
)

type Disconnect struct {
	ID PlayerID
}

func (d *Disconnect) Handle(addr *net.UDPAddr) error {
	if IsServer {
		log.Println("SERVER:", d.ID, "Has disonnceted")

		playerID := Lookup[addr.String()]
		var player *net.UDPAddr = Players[playerID]
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return nil
		}

		delete(Players, Lookup[addr.String()])
		delete(Lookup, addr.String())

		// tell others they left
		for _, p := range Players {
			Send(Disconnect{ID: playerID}, p)
		}
	} else {
		log.Println("Client:", Me, "--", d.ID, "Has disonnceted")
	}

	return nil
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
