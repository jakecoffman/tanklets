package tanklets

import (
	"log"
	"net"
	"github.com/jakecoffman/binser"
)

type Damage struct {
	ID PlayerID
	Killer PlayerID
}

func (d *Damage) Handle(addr *net.UDPAddr, game *Game) {
	tank := game.Tanks[d.ID]
	if tank == nil {
		log.Println("Tank", d.ID, "not found")
		return
	}
	tank.Destroyed = true

	if d.ID == Me {
		game.State = GameStateDead
	}
}

func (d Damage) MarshalBinary() ([]byte, error) {
	return d.Serialize(nil)
}

func (d *Damage) UnmarshalBinary(b []byte) error {
	_, err := d.Serialize(b)
	return err
}

func (d *Damage) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = DAMAGE
	stream.Uint8(&m)
	stream.Uint16((*uint16)(&d.ID))
	stream.Uint16((*uint16)(&d.Killer))
	return stream.Bytes()
}
