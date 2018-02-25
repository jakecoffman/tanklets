package server

import (
	"encoding"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/gutils"
	"github.com/jakecoffman/tanklets/pkt"
)

type Server struct {
	*tanklets.Net
	ServerAddr *net.UDPAddr
	UdpConn *net.UDPConn
}

func NewServer(addr string) *Server {
	ServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Init server connection")
	UdpConn, err := net.ListenUDP("udp", ServerAddr)
	if err != nil {
		log.Fatal(err)
	}

	UdpConn.SetReadBuffer(1048576)

	network := tanklets.NewNet()
	go network.Tick(func(){
		if network.NetworkIn > 0 && network.NetworkOut > 0 {
			fmt.Println("in :", gutils.Bytes(network.NetworkIn), "out:", gutils.Bytes(network.NetworkOut))
		}
	})

	return &Server{
		Net: network,
		ServerAddr: ServerAddr,
		UdpConn: UdpConn,
	}
}

func (s *Server) Close() error {
	s.Net.Close()
	return s.UdpConn.Close()
}

func (s *Server) Recv() {
	var addr *net.UDPAddr
	var err error
	var n int
	for {
		data := make([]byte, 1024)
		n, addr, err = s.UdpConn.ReadFromUDP(data)
		if err != nil {
			log.Println(err)
			return
		}
		atomic.AddUint64(&s.InBps, uint64(n))

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

		s.IncomingPackets <- tanklets.Packet{data, addr}
	}
}

func (s *Server) Send(handler encoding.BinaryMarshaler, addr *net.UDPAddr) {
	data, err := handler.MarshalBinary()
	if err != nil {
		log.Println(err)
		return
	}
	s.SendRaw(data, addr)
}

// SendRaw is the same as Send but takes bytes
func (s *Server) SendRaw(data []byte, addr *net.UDPAddr) {
	n, err := s.UdpConn.WriteToUDP(data, addr)
	if err != nil {
		panic(err)
		return
	}
	atomic.AddUint64(&s.OutBps, uint64(n))
}
