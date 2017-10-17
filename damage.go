package tanklets

import (
	"encoding/binary"
	"net"
)

type Damage struct {
	ID PlayerID
}

func (d *Damage) Handle(addr *net.UDPAddr) error {
	tank := Tanks[d.ID]
	tank.Destroyed = true

	if d.ID == Me {
		State = GAME_DEAD
	}

	return nil
}

func (d Damage) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 3)
	buf[0] = DAMAGE
	binary.BigEndian.PutUint16(buf[1:3], uint16(d.ID))
	return buf, nil
}

func (d *Damage) UnmarshalBinary(buf []byte) error {
	d.ID = PlayerID(binary.BigEndian.Uint16(buf[1:3]))
	return nil
}
