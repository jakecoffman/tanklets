package server

import (
	"net"
	"log"
	"sync/atomic"
	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
)

func Recv() {
	var addr *net.UDPAddr
	var err error
	var n int
	for {
		data := make([]byte, 1024)
		n, addr, err = tanklets.UdpConn.ReadFromUDP(data)
		if err != nil {
			log.Println(err)
			return
		}
		atomic.AddUint64(&tanklets.InBps, uint64(n))

		// handle certain things right now:
		switch data[0] {
		case pkt.PacketPing:
			ping := &pkt.Ping{}
			_, err := ping.Serialize(data)
			if err != nil {
				log.Println(err)
				continue
			}
			// TODO
			continue
		}

		tanklets.IncomingPackets <- tanklets.Packet{data, addr}
	}
}
