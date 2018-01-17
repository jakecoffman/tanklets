package tanklets

import (
	"bytes"
	"net"
	"time"
)

// Client's ping in ns
var MyPing time.Duration

type Ping struct {
	T time.Time
}

func (d *Ping) Handle(addr *net.UDPAddr) {
	if IsServer {
		tank := Tanks[Lookup[addr.String()]]
		tank.Ping = time.Now().Sub(d.T)
	} else {
		MyPing = time.Now().Sub(d.T)
		d.T = time.Now()
		ClientSend(d)
	}
}

func (d Ping) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{PING})
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
