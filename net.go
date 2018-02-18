package tanklets

import (
	"encoding"
	"log"
	"net"
	"sync/atomic"
	"time"
	"fmt"
	"github.com/jakecoffman/tanklets/gutils"
	"github.com/jakecoffman/tanklets/pkt"
)

var SimulatedNetworkLatencyMS = 100

var ServerAddr *net.UDPAddr
var UdpConn *net.UDPConn
var IsServer bool

type Packet struct {
	Bytes []byte
	Addr    *net.UDPAddr
}

type Outgoing struct {
	data []byte
	addr *net.UDPAddr
}

var IncomingPackets = make(chan Packet, 1000)

var tick = time.Tick(1 * time.Second)
var InBps uint64
var OutBps uint64

var NetworkIn, NetworkOut uint64

func init() {
	go func() {
		for {
			select {
			case <-tick:
				NetworkIn = atomic.LoadUint64(&InBps)
				NetworkOut = atomic.LoadUint64(&OutBps)
				atomic.StoreUint64(&InBps, 0)
				atomic.StoreUint64(&OutBps, 0)

				if IsServer {
					fmt.Println("in :", gutils.Bytes(NetworkIn))
					fmt.Println("out:", gutils.Bytes(NetworkOut))
				}
			}
		}
	}()
}

var ClientIsConnected, ClientIsConnecting bool

func NetInit(addr string) error {
	var err error

	ServerAddr, err = net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if IsServer {
		fmt.Println("Init server connection")
		UdpConn, err = net.ListenUDP("udp", ServerAddr)
		if err != nil {
			log.Fatal(err)
			return err
		}
	} else {
		ClientIsConnected = false
		ClientIsConnecting = true
		fmt.Println("Init client connection")
		UdpConn, err = net.DialUDP("udp", nil, ServerAddr)
		if err != nil {
			ClientIsConnecting = false
			log.Println(err)
			return err
		}

		defer ClientSend(pkt.Initial{})
	}

	UdpConn.SetReadBuffer(1048576)

	return nil
}

func NetClose() error {
	fmt.Println("Net close")
	ClientIsConnected = false
	return UdpConn.Close()
}

func ClientSend(handler encoding.BinaryMarshaler) {
	data, err := handler.MarshalBinary()
	if err != nil {
		log.Println(err)
		return
	}
	ClientSendRaw(data)
}

func ClientSendRaw(data []byte) {
	n, err := UdpConn.Write(data)
	if err != nil {
		panic(err)
		return
	}
	atomic.AddUint64(&OutBps, uint64(n))
}

func ServerSend(handler encoding.BinaryMarshaler, addr *net.UDPAddr) {
	data, err := handler.MarshalBinary()
	if err != nil {
		log.Println(err)
		return
	}
	ServerSendRaw(data, addr)
}

// SendRaw is the same as Send but takes bytes
func ServerSendRaw(data []byte, addr *net.UDPAddr) {
	n, err := UdpConn.WriteToUDP(data, addr)
	if err != nil {
		panic(err)
		return
	}
	atomic.AddUint64(&OutBps, uint64(n))
}

type Handler interface {
	Serialize(b []byte) ([]byte, error)
}
