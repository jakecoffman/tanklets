package pkt

import (
	"bytes"
	"time"
	"github.com/jakecoffman/binser"
)

// Client's ping in ns
var MyPing time.Duration

type Ping struct {
	T time.Time
}

func (d Ping) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{PacketPing})
	b, err := d.T.MarshalBinary()
	if err != nil {
		return nil, err
	}
	n, err := buf.Write(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes()[:n+1], nil
}

func (d *Ping) UnmarshalBinary(buf []byte) error {
	return d.T.UnmarshalBinary(buf[1:16])
}

func (d *Ping) Serialize(buf []byte) ([]byte, error) {
	stream := binser.NewStream(buf)
	var m uint8 = PacketPing
	stream.Uint8(&m)
	if !stream.IsReading() {
		b, err := d.T.MarshalBinary()
		if err != nil {
			return nil, err
		}
		stream.WriteBytes(b)
		return stream.Bytes()
	} else {
		return nil, d.T.UnmarshalBinary(buf[1:16])
	}
}
