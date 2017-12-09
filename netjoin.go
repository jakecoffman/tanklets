package tanklets

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binserializer"
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

func (j *Join) Handle(addr *net.UDPAddr) {
	var tank *Tank

	if IsServer {
		fmt.Println("Handling join")
		tank = NewTank(curId, colors[colorCursor])
		tank.SetPosition(cp.Vector{10 + float64(rand.Intn(400)), 10 + float64(rand.Intn(400))})
		// tell this player their ID
		Send(Join{tank.ID, true, tank.Color}, addr)
		loc := tank.Location()
		// tell this player where they are
		Send(loc, addr)
		join := Join{tank.ID, false, tank.Color}
		Players.Each(func (id PlayerID, p *net.UDPAddr) {
			// tell all players about this player
			Send(join, p)
			Send(loc, p)
			// tell this player where all the existing players are
			thisTank := Tanks[id]
			Send(Join{id, false, thisTank.Color}, addr)
			Send(thisTank.Location(), addr)
		})
		Lookup[addr.String()] = tank.ID
		Players.Put(curId, addr)
		curId++
		colorCursor++
	} else {
		fmt.Println("Player joined")
		tank = NewTank(j.ID, j.Color)
		if j.You {
			fmt.Println("Oh, it's me!")
			Me = tank.ID
			//Player = player
			State = GAME_PLAYING
		}
	}
	Tanks[tank.ID] = tank
}

func (j Join) MarshalBinary() ([]byte, error) {
	buf := binserializer.NewBuffer(17)
	buf.WriteByte(JOIN)
	if j.You {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	buf.WriteUint16(uint16(j.ID))
	buf.WriteFloat32(j.Color.X())
	buf.WriteFloat32(j.Color.Y())
	buf.WriteFloat32(j.Color.Z())
	return buf.Bytes()
}

func (j *Join) UnmarshalBinary(bytes []byte) error {
	buf := binserializer.NewBufferFromBytes(bytes)
	_ = buf.GetByte()
	if buf.GetByte() == 1 {
		j.You = true
	}
	j.ID = PlayerID(buf.GetUint16())
	j.Color[0] = buf.GetFloat32()
	j.Color[1] = buf.GetFloat32()
	j.Color[2] = buf.GetFloat32()
	return buf.Error()
}
