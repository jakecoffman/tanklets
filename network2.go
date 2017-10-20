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
	Handle(addr *net.UDPAddr)
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

func Marshal(fields []interface{}, buf *bytes.Buffer) ([]byte, error) {
	var err error
	for _, field := range fields {
		err = binary.Write(buf, binary.LittleEndian, field)
		if err != nil {
			log.Println(err)
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
