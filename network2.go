package tanklets

import (
	"bufio"
	"bytes"
	"encoding"
	"encoding/binary"
	"log"
	"net"
	"time"
)

const SimulatedNetworkLatencyMS = 100

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

var Incomings chan Incoming
var Outgoings chan Outgoing

func init() {
	ServerAddr = &net.UDPAddr{
		Port: 1234,
		IP:   net.ParseIP("127.0.0.1"),
	}
	Incomings = make(chan Incoming, 1000)
	Outgoings = make(chan Outgoing, 1000)
}

func NetInit() {
	var err error

	if IsServer {
		log.Println("Init server connection")
		udpConn, err = net.ListenUDP("udp", ServerAddr)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("Init client connection")
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
	log.Println("Net close")
	return udpConn.Close()
}

var incomingTick = time.Tick(1 * time.Second)
var incomingBytesPerSecond int
var outgoingTick = time.Tick(1 * time.Second)
var outgoingBytesPerSecond int

// Recv runs in a goroutine and un-marshals incoming data, queuing it up for ProcessIncoming to handle
func Recv() {
	for {
		data := make([]byte, 2048)
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
		incomingBytesPerSecond += n

		select {
		case <-incomingTick:
			if IsServer {
				log.Println("server in :", Bytes(incomingBytesPerSecond))
			} else {
				//log.Println("incoming client bytes:", incomingBytesPerSecond)
			}
			incomingBytesPerSecond = 0
		default:
		}

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
		default:
			log.Println("Unkown message type", data[0])
			continue
		}
		err = handler.UnmarshalBinary(data)
		if err != nil {
			log.Println(err)
			continue
		}
		//go func(){
		//	time.Sleep(SimulatedNetworkLatencyMS/2 * time.Millisecond)
		select {
		case Incomings <- Incoming{handler, addr}:
		default:
			// the idea here is the first message is probably out of date, so drop that one
			<-Incomings
			Incomings <- Incoming{handler, addr}
			log.Println("Error: queue is full, dropping message")
		}
		//}()
	}
}

// ProcessIncoming handles all incoming messages that are queued
func ProcessIncoming() {
	var err error
	for {
		select {
		case incoming := <-Incomings:
			if err = incoming.Handler.Handle(incoming.Addr); err != nil {
				log.Fatal(err)
			}
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

	//go func() {
	//	time.Sleep(SimulatedNetworkLatencyMS / 2 * time.Millisecond)
	Outgoings <- Outgoing{data: data, addr: addr}
	//}()
}

func ProcessOutgoingServer() {
	var outgoing Outgoing
	for {
		outgoing = <-Outgoings
		n, err := udpConn.WriteToUDP(outgoing.data, outgoing.addr)
		if err != nil {
			panic(err)
			return
		}
		outgoingBytesPerSecond += n
		select {
		case <-outgoingTick:
			log.Println("server out:", Bytes(outgoingBytesPerSecond))
			outgoingBytesPerSecond = 0
		default:
		}
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
		outgoingBytesPerSecond += n
		select {
		case <-outgoingTick:
			//log.Println("outgoing client bytes:", outgoingBytesPerSecond)
			outgoingBytesPerSecond = 0
		default:
		}
	}
}

type Handler interface {
	Handle(addr *net.UDPAddr) error
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

func Marshal(fields []interface{}, buf *bytes.Buffer) ([]byte, error) {
	var err error
	for _, field := range fields {
		err = binary.Write(buf, binary.LittleEndian, field)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func Unmarshal(fields []interface{}, reader *bytes.Reader) error {
	for _, field := range fields {
		err := binary.Read(reader, binary.LittleEndian, field)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
