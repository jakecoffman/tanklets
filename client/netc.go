package client

import (
	"bufio"
	"log"
	"sync/atomic"
	"github.com/jakecoffman/tanklets"
	"time"
	"fmt"
	"github.com/jakecoffman/tanklets/pkt"
)

func Recv() {
	var err error
	var n int
	for {
		data := make([]byte, 2048)
		n, err = bufio.NewReader(tanklets.UdpConn).Read(data)
		if err != nil {
			tanklets.ClientIsConnected = false
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
			pkt.MyPing = time.Now().Sub(ping.T)
			ping.T = time.Now()
			tanklets.ClientSend(ping)
			continue
		}

		if data[0] == pkt.PacketJoin {
			fmt.Println("CLIENT QUEUING UP JOIN")
		}

		tanklets.IncomingPackets <- tanklets.Packet{data, nil}
	}
}
