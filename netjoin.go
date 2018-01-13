package tanklets

import (
	"fmt"
	"math/rand"
	"net"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binser"
	"golang.org/x/image/math/f32"
)

var curId PlayerID = 1

type Join struct {
	ID    PlayerID
	You   uint8
	Color f32.Vec3
}

var colors = []mgl32.Vec3{
	{1, 0, 0},
	{0, 1, 0},
	{0, 0, 1},
	{1, 1, 0},
	{0, 1, 1},
	{1, 0, 1},
	{1, 1, 1},
	{.5, 0, 0},
	{0, .5, 0},
	{0, 0, .5},
	{.5, .5, 0},
	{0, .5, .5},
	{.5, 0, .5},
	{.5, .5, .5},
}

var colorCursor int

func (j *Join) Handle(addr *net.UDPAddr) {
	var tank *Tank

	if IsServer {
		fmt.Println("Handling join")
		tank = NewTank(curId, colors[colorCursor])
		tank.SetPosition(cp.Vector{10 + float64(rand.Intn(400)), 10 + float64(rand.Intn(400))})
		// tell this player their ID
		Send(Join{tank.ID, 1, f32.Vec3(tank.Color)}, addr)
		loc := tank.Location()
		// tell this player where they are
		Send(loc, addr)
		join := Join{tank.ID, 0, f32.Vec3(tank.Color)}
		Players.Each(func (id PlayerID, p *net.UDPAddr) {
			// tell all players about this player
			Send(join, p)
			Send(loc, p)
			// tell this player where all the existing players are
			thisTank := Tanks[id]
			Send(Join{id, 0, f32.Vec3(thisTank.Color)}, addr)
			Send(thisTank.Location(), addr)
		})
		Lookup[addr.String()] = tank.ID
		Players.Put(curId, addr)
		curId++
		colorCursor++
	} else {
		fmt.Println("Player joined")
		tank = NewTank(j.ID, mgl32.Vec3(j.Color))
		if j.You > 0 {
			fmt.Println("Oh, it's me!")
			Me = tank.ID
			//Player = player
			State = GAME_PLAYING
		}
	}
	Tanks[tank.ID] = tank
}

func (j Join) MarshalBinary() ([]byte, error) {
	return j.Serialize(nil)
}

func (j *Join) UnmarshalBinary(b []byte) error {
	_, err := j.Serialize(b)
	return err
}

func (j *Join) Serialize(b []byte) ([]byte, error) {
	stream := binser.NewStream(b)
	var m uint8 = JOIN
	stream.Uint8(&m)
	stream.Uint8(&j.You)
	stream.Uint16((*uint16)(&j.ID))
	stream.Float32(&j.Color[0])
	stream.Float32(&j.Color[1])
	stream.Float32(&j.Color[2])
	return stream.Bytes()
}
