package server

import (
	"github.com/jakecoffman/tanklets"
	"fmt"
	"github.com/jakecoffman/cp"
	"net"
	"math/rand"
	"golang.org/x/image/math/f32"
	"log"
	"time"
	"github.com/jakecoffman/tanklets/pkt"
)

type packetHandler func(packet tanklets.Packet, game *tanklets.Game)

var handlers [pkt.PacketMax]packetHandler

func init() {
	for i := 0; i < pkt.PacketMax; i++ {
		handlers[i] = noop
	}

	handlers[pkt.PacketInit] = initial
	handlers[pkt.PacketJoin] = join
	handlers[pkt.PacketDisconnect] = disconnect
	handlers[pkt.PacketMove] = move
	handlers[pkt.PacketPing] = ping
	handlers[pkt.PacketReady] = ready
	handlers[pkt.PacketShoot] = shoot
}

func ProcessNetwork(packet tanklets.Packet, game *tanklets.Game) {
	handlers[packet.Bytes[0]](packet, game)
}

func noop(packet tanklets.Packet, _ *tanklets.Game) {
	log.Println("Unhandled server packet", packet.Bytes[0])
}

func initial(packet tanklets.Packet, game *tanklets.Game) {
	addr := packet.Addr
	initial := pkt.Initial{}
	_, err := initial.Serialize(packet.Bytes)
	if err != nil {
		log.Println(err)
		return
	}
	id, ok := tanklets.Lookup[addr.String()]

	if ok {
		initial.ID = id
		fmt.Println("Player", id, "reconnected", addr)
	} else {
		id = tanklets.PlayerID(game.CursorPlayerId.Next())
		initial.ID = id
		tanklets.Lookup[addr.String()] = id
		tanklets.Players.Put(id, addr)
		fmt.Println("Player", id, "connected", addr)
	}

	tanklets.ServerSend(initial, addr)
}

func join(packet tanklets.Packet, game *tanklets.Game) {
	fmt.Println("SERVER Handling join")
	addr := packet.Addr
	tank := game.NewTank(tanklets.Lookup[addr.String()], pkt.GetColor(game.CursorColor.Next()))
	tank.SetPosition(cp.Vector{10 + float64(rand.Intn(790)), 10 + float64(rand.Intn(580))})
	// tell this player their ID
	tanklets.ServerSend(pkt.Join{tank.ID, 1, f32.Vec3(tank.Color)}, addr)
	loc := tank.Location()
	// tell this player where they are
	tanklets.ServerSend(loc, addr)
	join := pkt.Join{tank.ID, 0, f32.Vec3(tank.Color)}
	tanklets.Players.Each(func (id tanklets.PlayerID, p *net.UDPAddr) {
		if id == tank.ID {
			return
		}
		// tell all players about this player
		tanklets.ServerSend(join, p)
		tanklets.ServerSend(loc, p)
		// tell this player where all the existing players are
		thisTank := game.Tanks[id]
		tanklets.ServerSend(pkt.Join{id, 0, f32.Vec3(thisTank.Color)}, addr)
		tanklets.ServerSend(thisTank.Location(), addr)
	})
	// Tell this player about the level
	for _, box := range game.Boxes {
		tanklets.ServerSend(box.Location(), addr)
	}
	tanklets.Lookup[addr.String()] = tank.ID
	tanklets.Players.Put(tank.ID, addr)
	game.Tanks[tank.ID] = tank
	fmt.Println("tank", tank.ID, "joined")
}

func disconnect(packet tanklets.Packet, game *tanklets.Game) {
	addr := packet.Addr

	playerID := tanklets.Lookup[addr.String()]
	player := tanklets.Players.Get(playerID)
	if player == nil {
		// this is normal, we spam disconnect when leaving to ensure the server gets it
		return
	}

	tanklets.Players.Delete(playerID)
	delete(tanklets.Lookup, addr.String())
	game.Tanks[playerID].Destroyed = true

	// tell others they left & destroyed
	tanklets.Players.SendAll(pkt.Disconnect{ID: uint16(playerID)}, pkt.Damage{ID: playerID, Killer: playerID})
}
func move(packet tanklets.Packet, game *tanklets.Game) {
	m := pkt.Move{}
	if _, err := m.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}
	addr := packet.Addr

	if game.State != tanklets.GameStatePlaying {
		return
	}

	tank := game.Tanks[tanklets.Lookup[addr.String()]]
	if tank == nil {
		log.Println("Player not found", addr.String(), tanklets.Lookup[addr.String()])
		return
	}
	if tank.Destroyed {
		return
	}

	tank.NextMove.Turn = m.Turn
	tank.NextMove.Throttle = m.Throttle
	tank.NextMove.TurretAngle = m.TurretAngle
}
func ping(packet tanklets.Packet, game *tanklets.Game) {}
func ready(packet tanklets.Packet, game *tanklets.Game) {
	tank := game.Tanks[tanklets.Lookup[packet.Addr.String()]]
	tank.Ready = true
}
func shoot(packet tanklets.Packet, game *tanklets.Game) {
	addr := packet.Addr
	id := tanklets.Lookup[addr.String()]
	player := tanklets.Players.Get(id)
	if player == nil {
		log.Println("Player not found", addr.String(), tanklets.Lookup[addr.String()])
		return
	}
	tank := game.Tanks[id]

	if time.Now().Sub(tank.LastShot) < tanklets.ShotCooldown {
		return
	}
	tank.LastShot = time.Now()

	bullet := game.NewBullet(tank, tanklets.BulletID(game.CursorBullet.Next()))

	pos := cp.Vector{X: tanklets.TankHeight / 2.0}
	pos = pos.Rotate(tank.Turret.Rotation())
	bullet.Body.SetPosition(pos.Add(tank.Turret.Position()))
	bullet.Body.SetAngle(tank.Turret.Angle())
	bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{tanklets.BulletSpeed, 0}))
	//bullet.Shape.SetFilter(cp.NewShapeFilter(uint(player.ID), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	shot := bullet.Location()
	tanklets.Players.SendAll(shot)
}
func state(packet tanklets.Packet, game *tanklets.Game) {}
