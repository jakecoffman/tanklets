package tanklets

import (
	"log"
	"math/rand"
	"net"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

var curId PlayerID = 1

type Join struct {
	ID  PlayerID
	You bool
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
	colorCursor++

	if IsServer {
		// player initialization (TODO set spawn point)
		player = NewTank(curId, colors[colorCursor])
		player.SetPosition(cp.Vector{10 + float64(rand.Intn(400)), 10 + float64(rand.Intn(400))})
		player.Addr = addr
		curId++
		Lookup[addr.String()] = player.ID
		// tell this player their ID
		b, err := (&Join{ID: player.ID, You: true}).MarshalBinary()
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
		joinBytes, err := Join{player.ID, false}.MarshalBinary()
		if err != nil {
			log.Println(err)
			return err
		}
		for _, p := range Tanks {
			// tell all players about this player
			Send(joinBytes, p.Addr)
			Send(loc, p.Addr)
			// tell this player where all the existing players are
			b, err = (&Join{p.ID, false}).MarshalBinary()
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
		player = NewTank(j.ID, colors[colorCursor])
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
	if j.You {
		return []byte{JOIN, byte(j.ID), 1}, nil
	} else {
		return []byte{JOIN, byte(j.ID), 0}, nil
	}
}

func (j *Join) UnmarshalBinary(b []byte) error {
	j.ID = PlayerID(b[1])
	if b[2] == 0 {
		j.You = false
	} else {
		j.You = true
	}
	return nil
}
