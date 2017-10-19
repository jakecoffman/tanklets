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
	{1, 0, 0},
	{0, 1, 0},
	{0, 0, 1},
	{1, 1, 0},
	{0, 1, 1},
	{1, 0, 1},
	{1, 1, 1},
}

var colorCursor int

func (j *Join) Handle(addr *net.UDPAddr) error {
	var tank *Tank

	if IsServer {
		tank = NewTank(curId, colors[colorCursor])
		tank.SetPosition(cp.Vector{10 + float64(rand.Intn(400)), 10 + float64(rand.Intn(400))})
		// tell this player their ID
		Send(Join{tank.ID, true, tank.Color}, addr)
		loc := tank.Location()
		// tell this player where they are
		Send(loc, addr)
		join := Join{tank.ID, false, tank.Color}
		for id, p := range Players {
			// tell all players about this player
			Send(join, p)
			Send(loc, p)
			// tell this player where all the existing players are
			thisTank := Tanks[id]
			Send(Join{id, false, thisTank.Color}, addr)
			Send(thisTank.Location(), addr)
		}
		Lookup[addr.String()] = tank.ID
		Players[curId] = addr
		curId++
		colorCursor++
	} else {
		log.Println("Player joined")
		tank = NewTank(j.ID, j.Color)
		if j.You {
			log.Println("Oh, it's me!")
			Me = tank.ID
			//Player = player
			State = GAME_PLAYING
		}
	}
	Tanks[tank.ID] = tank

	return nil
}

func (j Join) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 17)
	buf[0] = JOIN
	binary.BigEndian.PutUint16(buf[1:3], uint16(j.ID))
	if j.You {
		binary.BigEndian.PutUint16(buf[3:5], uint16(1))
	} else {
		binary.BigEndian.PutUint16(buf[3:5], uint16(0))
	}
	binary.BigEndian.PutUint32(buf[5:9], uint32(j.Color.X()))
	binary.BigEndian.PutUint32(buf[9:13], uint32(j.Color.Y()))
	binary.BigEndian.PutUint32(buf[13:17], uint32(j.Color.Z()))
	return buf, nil
}

func (j *Join) UnmarshalBinary(buf []byte) error {
	j.ID = PlayerID(binary.BigEndian.Uint16(buf[1:3]))
	j.You = binary.BigEndian.Uint16(buf[3:5]) == 1
	j.Color[0] = float32(binary.BigEndian.Uint32(buf[5:9]))
	j.Color[1] = float32(binary.BigEndian.Uint32(buf[9:13]))
	j.Color[2] = float32(binary.BigEndian.Uint32(buf[13:17]))
	return nil
}
