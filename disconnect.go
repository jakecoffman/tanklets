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

		var player *Tank = Tanks[Lookup[addr.String()]]
		if player == nil {
			log.Println("Player not found", addr.String(), Lookup[addr.String()])
			return nil
		}

		delete(Tanks, Lookup[addr.String()])
		delete(Lookup, addr.String())

		// tell others they left
		for _, p := range Tanks {
			Send(Disconnect{ID: player.ID}, p.Addr)
		}
	} else {
		log.Println("Client:", Me, "--", d.ID, "Has disonnceted")
		delete(Tanks, d.ID)
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
