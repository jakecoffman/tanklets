package tanklets

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

var curId PlayerID = 1

type Join struct {
	ID    PlayerID
	You   bool
	Color mgl32.Vec3
}

var colors = []mgl32.Vec3{
	{1, 1, 1},
	{1, 0, 0},
	{0, 1, 0},
	{0, 0, 1},
	{1, 1, 0},
	{0, 1, 1},
	{1, 0, 1},
}

var colorCursor int

func (j *Join) Handle(addr *net.UDPAddr) error {
	var player *Tank

	if IsServer {
		// player initialization (TODO set spawn point)
		player = NewTank(curId, colors[colorCursor])
		colorCursor++
		player.SetPosition(cp.Vector{10 + float64(rand.Intn(400)), 10 + float64(rand.Intn(400))})
		player.Addr = addr
		curId++
		Lookup[addr.String()] = player.ID
		// tell this player their ID
		b, err := (&Join{player.ID, true, player.Color}).MarshalBinary()
		if err != nil {
			log.Println(err)
			return err
		}
		Send(b, addr)
		loc, err := player.Location().MarshalBinary()
		if err != nil {
			log.Println(err)
			return err
		}
		// tell this player where they are
		Send(loc, addr)
		joinBytes, err := Join{player.ID, false, player.Color}.MarshalBinary()
		if err != nil {
			log.Println(err)
			return err
		}
		for _, p := range Tanks {
			// tell all players about this player
			Send(joinBytes, p.Addr)
			Send(loc, p.Addr)
			// tell this player where all the existing players are
			b, err = (&Join{p.ID, false, p.Color}).MarshalBinary()
			if err != nil {
				log.Println(err)
				continue
			}
			Send(b, player.Addr)
			b, err = p.Location().MarshalBinary()
			if err != nil {
				log.Println(err)
				continue
			}
			Send(b, player.Addr)
		}
	} else {
		log.Println("Player joined")
		player = NewTank(j.ID, j.Color)
		if j.You {
			log.Println("Oh, it's me!")
			Me = player.ID
			//Player = player
			State = GAME_ACTIVE
			// now that I am joined I will start pinging the server
			go PingRegularly()
		}
	}
	Tanks[player.ID] = player

	return nil
}

func (j Join) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint16(buf[0:2], uint16(j.ID))
	if j.You {
		binary.BigEndian.PutUint16(buf[2:4], uint16(1))
	} else {
		binary.BigEndian.PutUint16(buf[2:4], uint16(0))
	}
	binary.BigEndian.PutUint32(buf[4:8], uint32(j.Color.X()))
	binary.BigEndian.PutUint32(buf[8:12], uint32(j.Color.Y()))
	binary.BigEndian.PutUint32(buf[12:16], uint32(j.Color.Z()))
	return buf, nil
}

func (j *Join) UnmarshalBinary(buf []byte) error {
	j.ID = PlayerID(binary.BigEndian.Uint16(buf[0:2]))
	j.You = binary.BigEndian.Uint16(buf[2:4]) == 1
	j.Color[0] = float32(binary.BigEndian.Uint32(buf[4:8]))
	j.Color[1] = float32(binary.BigEndian.Uint32(buf[8:12]))
	j.Color[2] = float32(binary.BigEndian.Uint32(buf[12:16]))
	return nil
}
