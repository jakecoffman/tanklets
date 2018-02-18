package pkt

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/binser"
	"golang.org/x/image/math/f32"
)

type Join struct {
	ID    PlayerID
	You   uint8
	Color f32.Vec3
	Name  string
}

func GetColor(i int) mgl32.Vec3 {
	return []mgl32.Vec3{
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
	}[i]
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
	var m uint8 = PacketJoin
	stream.Uint8(&m)
	stream.Uint8(&j.You)
	stream.Uint16((*uint16)(&j.ID))
	stream.Float32(&j.Color[0])
	stream.Float32(&j.Color[1])
	stream.Float32(&j.Color[2])

	// TODO implement this better in binser
	var n uint8
	if !stream.IsReading() {
		n = uint8(len(j.Name))
		stream.Uint8(&n)
		stream.WriteBytes([]byte(j.Name))
	} else {
		stream.Uint8(&n)
		j.Name = string(stream.GetBytes(int(n)))
	}
	return stream.Bytes()
}
