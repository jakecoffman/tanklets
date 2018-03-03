package server

import (
	"encoding"
	"log"
	"net"
	"sync"

	"github.com/jakecoffman/tanklets/pkt"
)

type PlayerID = pkt.PlayerID

// PlayerLookup lets the server find players by their address or PlayerId
type PlayerLookup struct {
	sync.RWMutex
	players map[PlayerID]*net.UDPAddr
	lookup map[string]PlayerID
}

func (p *PlayerLookup) Get(id PlayerID) *net.UDPAddr {
	p.RLock()
	defer p.RUnlock()
	return p.players[id]
}

func (p *PlayerLookup) Lookup(addr string) (PlayerID, bool) {
	p.RLock()
	defer p.RUnlock()
	id, ok := p.lookup[addr]
	return id, ok
}

func (p *PlayerLookup) Put(id PlayerID, addr *net.UDPAddr) {
	p.Lock()
	p.players[id] = addr
	p.lookup[addr.String()] = id
	p.Unlock()
}

func (p *PlayerLookup) Delete(id PlayerID) {
	p.Lock()
	addr := p.players[id]
	delete(p.players, id)
	delete(p.lookup, addr.String())
	p.Unlock()
}

func (p *PlayerLookup) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.players)
}

func (p *PlayerLookup) Each(f func (PlayerID, *net.UDPAddr)) {
	p.RLock()
	defer p.RUnlock()
	for id, addr := range p.players {
		f(id, addr)
	}
}

func (p *PlayerLookup) SendAll(network *Server, packets ...encoding.BinaryMarshaler) {
	p.RLock()
	for _, packet := range packets {
		data, err := packet.MarshalBinary()
		if err != nil {
			log.Fatal(err)
		}
		for _, player := range p.players {
			network.SendRaw(data, player)
		}
	}
	p.RUnlock()
}