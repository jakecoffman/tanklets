package network

import (
	"encoding/binary"
	"fmt"
	"math"
)

type Thing struct {
	A int32
	B uint8
}

func (t *Thing) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 6)
	binary.BigEndian.PutUint32(buf[0:4], uint32(t.A))
	binary.BigEndian.PutUint16(buf[4:6], uint16(t.B))
	return buf, nil
}

func (t *Thing) UnmarshalBinary(buf []byte) error {
	t.A = int32(binary.BigEndian.Uint32(buf[0:4]))
	t.B = uint8(binary.BigEndian.Uint16(buf[4:6]))
	return nil
}

func main() {
	a := &Thing{A: -1, B: math.MaxUint8}
	b := &Thing{}

	bits, err := a.MarshalBinary()
	if err != nil {
		panic(err)
	}
	err = b.UnmarshalBinary(bits)
	if err != nil {
		panic(err)
	}

	if *a != *b {
		panic(fmt.Sprint("NEQ:", *a, *b))
	}
}
