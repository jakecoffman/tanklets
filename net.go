package tanklets

import (
	"bufio"
	"encoding"
	"log"
	"net"
	"sync/atomic"
	"time"
	"fmt"
	"github.com/jakecoffman/tanklets/gutils"
)

var SimulatedNetworkLatencyMS = 100

// Message type IDs
const (
	INIT        = iota
	JOIN
	DISCONNECT
	MOVE
	SHOOT
	LOCATION
	BOXLOCATION
	DAMAGE
	PING
)

var ServerAddr *net.UDPAddr
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
		udpConn, err = net.ListenUDP("udp", ServerAddr)
		if err != nil {
			log.Fatal(err)
			return err
		}
	} else {
		ClientIsConnected = false
		ClientIsConnecting = true
		fmt.Println("Init client connection")
		udpConn, err = net.DialUDP("udp", nil, ServerAddr)
		if err != nil {
			ClientIsConnecting = false
			log.Println(err)
			return err
		}

		defer ClientSend(Init{})
	}

	udpConn.SetReadBuffer(1048576)

	go Recv()
	return nil
}

func NetClose() error {
	fmt.Println("Net close")
	ClientIsConnected = false
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
				ClientIsConnected = false
				log.Println(err)
				return
			}
		}
		atomic.AddUint64(&incomingBytesPerSecond, uint64(n))

		var handler Handler
		switch data[0] {
		case INIT:
			if !IsServer {
				init := &Init{}
				init.Handle(addr, nil)
				continue
			} else {
				handler = &Init{}
			}
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
		case BOXLOCATION:
			handler = &BoxLocation{}
		case DAMAGE:
			handler = &Damage{}
		case PING:
			handler = &Ping{}
			_, err = handler.Serialize(data)
			if err != nil {
				log.Println(err)
				continue
			}
			handler.Handle(addr, nil)
			continue
		default:
			log.Println("Unkown message type", data[0])
			continue
		}
		_, err = handler.Serialize(data)
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
func ProcessIncoming(game *Game) {
	for {
		select {
		case incoming := <-Incomings:
			incoming.Handler.Handle(incoming.Addr, game)
		default:
			// no data to process this frame
			return
		}
	}
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
	n, err := udpConn.Write(data)
	if err != nil {
		panic(err)
		return
	}
	atomic.AddUint64(&outgoingBytesPerSecond, uint64(n))
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
	n, err := udpConn.WriteToUDP(data, addr)
	if err != nil {
		panic(err)
		return
	}
	atomic.AddUint64(&outgoingBytesPerSecond, uint64(n))
}

type Handler interface {
	Handle(addr *net.UDPAddr, game *Game)
	Serialize(b []byte) ([]byte, error)
}
