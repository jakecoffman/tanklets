package tanklets

import (
	"bufio"
	"encoding"
	"log"
	"net"
	"sync/atomic"
	"time"
	"fmt"
)

var SimulatedNetworkLatencyMS = 100

// Message type IDs
const (
	JOIN = iota
	DISCONNECT
	MOVE
	SHOOT
	LOCATION
	DAMAGE
	PING
)

var ServerAddr = &net.UDPAddr{
	Port: 1234,
	IP:   net.ParseIP("127.0.0.1"),
}
var udpConn *net.UDPConn
var IsServer bool

type Incoming struct {
	Handler Handler
	Addr    *net.UDPAddr
}

type Outgoing struct {
	data []byte
	addr *net.UDPAddr
}

var Incomings = make(chan Incoming, 1000)
var Outgoings = make(chan Outgoing, 1000)

var tick = time.Tick(1 * time.Second)
var incomingBytesPerSecond uint64
var outgoingBytesPerSecond uint64

var NetworkIn, NetworkOut uint64

func init() {
	go func() {
		for {
			select {
			case <-tick:
				NetworkIn = atomic.LoadUint64(&incomingBytesPerSecond)
				NetworkOut = atomic.LoadUint64(&outgoingBytesPerSecond)
				atomic.StoreUint64(&incomingBytesPerSecond, 0)
				atomic.StoreUint64(&outgoingBytesPerSecond, 0)

				if IsServer {
					fmt.Println("in :", Bytes(NetworkIn))
					fmt.Println("out:", Bytes(NetworkOut))
				}
			}
		}
	}()
}

func NetInit() {
	var err error

	if IsServer {
		fmt.Println("Init server connection")
		udpConn, err = net.ListenUDP("udp", ServerAddr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Init client connection")
		udpConn, err = net.DialUDP("udp", nil, ServerAddr)
		if err != nil {
			log.Fatal(err)
		}
	}

	udpConn.SetReadBuffer(1048576)

	go Recv()

	if IsServer {
		go ProcessOutgoingServer()
	} else {
		go ProcessingOutgoingClient()
	}
}

func NetClose() error {
	fmt.Println("Net close")
	return udpConn.Close()
}

// Recv runs in a goroutine and un-marshals incoming data, queuing it up for ProcessIncoming to handle
func Recv() {
	data := make([]byte, 2048)
	for {
		var addr *net.UDPAddr
		var err error
		var n int
		if IsServer {
			n, addr, err = udpConn.ReadFromUDP(data)
			if err != nil {
				panic(err)
				return
			}
		} else {
			n, err = bufio.NewReader(udpConn).Read(data)
			if err != nil {
				log.Println(err)
				return
			}
		}
		atomic.AddUint64(&incomingBytesPerSecond, uint64(n))

		var handler Handler
		switch data[0] {
		case JOIN:
			handler = &Join{}
		case DISCONNECT:
			handler = &Disconnect{}
		case MOVE:
			handler = &Move{}
		case SHOOT:
			handler = &Shoot{}
		case LOCATION:
			handler = &Location{}
		case DAMAGE:
			handler = &Damage{}
		case PING:
			handler := &Ping{}
			// just handle ping right now
			err = handler.UnmarshalBinary(data)
			if err != nil {
				log.Println(err)
				continue
			}
			handler.Handle(addr)
			continue
		default:
			log.Println("Unkown message type", data[0])
			continue
		}
		err = handler.UnmarshalBinary(data)
		if err != nil {
			log.Println(err)
			continue
		}
		incoming := Incoming{handler, addr}
		select {
		case Incomings <- incoming:
		default:
			// the first message is more likely to be out of date, so drop that one
			<-Incomings
			Incomings <- incoming
			log.Println("Error: queue is full, dropping message")
		}
	}
}

// ProcessIncoming runs on the game thread and handles all incoming messages that are queued
func ProcessIncoming() {
	for {
		select {
		case incoming := <-Incomings:
			incoming.Handler.Handle(incoming.Addr)
		default:
			// no data to process this frame
			return
		}
	}
}

// Send queues up an outgoing byte array to be sent immediately so sending isn't blocking
func Send(handler encoding.BinaryMarshaler, addr *net.UDPAddr) {
	data, err := handler.MarshalBinary()
	if err != nil {
		log.Println(err)
		return
	}

	Outgoings <- Outgoing{data: data, addr: addr}
}

// SendRaw is the same as Send but takes bytes
func SendRaw(data []byte, addr *net.UDPAddr) {
	Outgoings <- Outgoing{data: data, addr: addr}
}

func ProcessOutgoingServer() {
	var outgoing Outgoing
	var n int
	var err error

	for {
		outgoing = <-Outgoings
		n, err = udpConn.WriteToUDP(outgoing.data, outgoing.addr)
		if err != nil {
			panic(err)
			return
		}
		atomic.AddUint64(&outgoingBytesPerSecond, uint64(n))
	}
}

func ProcessingOutgoingClient() {
	var outgoing Outgoing
	var n int
	var err error
	for {
		outgoing = <-Outgoings
		n, err = udpConn.Write(outgoing.data)
		if err != nil {
			log.Println(err)
		}
		atomic.AddUint64(&outgoingBytesPerSecond, uint64(n))
	}
}

type Handler interface {
	Handle(addr *net.UDPAddr)
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}
