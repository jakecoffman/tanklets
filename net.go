package tanklets

import (
	"net"
	"sync/atomic"
	"time"
)

type Packet struct {
	Bytes []byte
	Addr    *net.UDPAddr
}

type Net struct {
	ticker *time.Ticker
	stop chan struct{}

	InBps, OutBps uint64
	NetworkIn, NetworkOut uint64

	IncomingPackets chan Packet
}

func NewNet() *Net {
	return &Net{
		IncomingPackets: make(chan Packet, 1000),
		stop: make(chan struct{}),
	}
}

func (n *Net) Close() {
	n.stop<- struct{}{}
}

func (n *Net) Tick() {
	n.ticker = time.NewTicker(1 * time.Second)
	defer n.ticker.Stop()

	for {
		select {
		case <-n.ticker.C:
			n.NetworkIn = atomic.LoadUint64(&n.InBps)
			n.NetworkOut = atomic.LoadUint64(&n.OutBps)
			atomic.StoreUint64(&n.InBps, 0)
			atomic.StoreUint64(&n.OutBps, 0)

			//if IsServer && n.NetworkIn > 0 && n.NetworkOut > 0 {
			//	fmt.Println("in :", gutils.Bytes(n.NetworkIn), "out:", gutils.Bytes(n.NetworkOut))
			//}
		case <-n.stop:
			close(n.stop)
			return
		}
	}
}
