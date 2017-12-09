package tanklets

import (
	"log"
	"net"
	"github.com/jakecoffman/binserializer"
)

type Damage struct {
	ID PlayerID
}

func (d *Damage) Handle(addr *net.UDPAddr) {
	tank := Tanks[d.ID]
	if tank == nil {
		log.Println("Tank", d.ID, "not found")
		return
	}
	tank.Destroyed = true

	if d.ID == Me {
		State = GAME_DEAD
	}
}

func (d Damage) MarshalBinary() ([]byte, error) {
	buf := binserializer.NewBuffer(3)
	buf.WriteByte(DAMAGE)
	buf.WriteUint16(uint16(d.ID))
	return buf.Bytes()
}

func (d *Damage) UnmarshalBinary(b []byte) error {
	buf := binserializer.NewBufferFromBytes(b)
	_ = buf.GetByte()
	d.ID = PlayerID(buf.GetUint16())
	return nil
}
