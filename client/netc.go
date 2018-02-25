package client

import (
	"bufio"
	"encoding"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
)

type Client struct {
	*tanklets.Net
	ServerAddr *net.UDPAddr
	UdpConn *net.UDPConn

	IsConnecting, IsConnected bool
}

func NewClient(addr string) (*Client, error) {
	ServerAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Init client connection")
	UdpConn, err := net.DialUDP("udp", nil, ServerAddr)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Net: tanklets.NewNet(),
		ServerAddr: ServerAddr,
		UdpConn: UdpConn,
		IsConnecting: true,
	}

	UdpConn.SetReadBuffer(1048576)
	client.Send(pkt.Initial{})

	go client.Tick()

	return client, nil
}

func (c *Client) Close() error {
	c.Net.Close()
	c.IsConnected = false
	return c.UdpConn.Close()
}

func (c *Client) Recv() {
	var err error
	var n int
	for {
		data := make([]byte, 2048)
		n, err = bufio.NewReader(c.UdpConn).Read(data)
		if err != nil {
			c.IsConnected = false
			log.Println(err)
			return
		}
		atomic.AddUint64(&c.InBps, uint64(n))

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
			c.Send(ping)
			continue
		}

		if data[0] == pkt.PacketJoin {
			fmt.Println("CLIENT QUEUING UP JOIN")
		}

		c.IncomingPackets <- tanklets.Packet{data, nil}
	}
}

func (c *Client) Send(handler encoding.BinaryMarshaler) {
	data, err := handler.MarshalBinary()
	if err != nil {
		log.Println(err)
		return
	}
	c.SendRaw(data)
}

func (c *Client) SendRaw(data []byte) {
	n, err := c.UdpConn.Write(data)
	if err != nil {
		panic(err)
		return
	}
	atomic.AddUint64(&c.OutBps, uint64(n))
}
