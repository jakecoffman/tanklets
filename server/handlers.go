package server

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/pkt"
	"golang.org/x/image/math/f32"
)

type packetHandler func(packet tanklets.Packet, game *Game)

var handlers [pkt.PacketMax]packetHandler

func init() {
	for i := 0; i < pkt.PacketMax; i++ {
		handlers[i] = noop
	}

	handlers[pkt.PacketInit] = initial
	handlers[pkt.PacketJoin] = join
	handlers[pkt.PacketDisconnect] = disconnect
	handlers[pkt.PacketMove] = move
	handlers[pkt.PacketReady] = ready
	handlers[pkt.PacketShoot] = shoot
	handlers[pkt.PacketPing] = ping
}

func ProcessNetwork(packet tanklets.Packet, game *Game) {
	handlers[packet.Bytes[0]](packet, game)
}

func noop(packet tanklets.Packet, _ *Game) {
	log.Println("Unhandled server packet", packet.Bytes[0])
}

func initial(packet tanklets.Packet, game *Game) {
	addr := packet.Addr
	initial := pkt.Initial{}
	_, err := initial.Serialize(packet.Bytes)
	if err != nil {
		log.Println(err)
		return
	}
	id, ok := Players.Lookup(addr.String())

	if ok {
		initial.ID = id
		fmt.Println("Player", id, "reconnected", addr)
	} else {
		id = tanklets.PlayerID(game.CursorPlayerId.Next())
		initial.ID = id
		Players.Put(id, addr)
		fmt.Println("Player", id, "connected", addr)
	}

	game.Network.Send(initial, addr)
}

var HasHadPlayersConnect bool

func join(packet tanklets.Packet, game *Game) {
	addr := packet.Addr
	playerId, _ := Players.Lookup(addr.String())
	tank := game.Tanks[playerId]

	fmt.Println("Processing JOIN")

	if tank != nil {
		tank.LastPkt = time.Now()
		// they are already here, so this is a rejoin or name change
		j := pkt.Join{}
		if _, err := j.Serialize(packet.Bytes); err != nil {
			log.Println(err)
			return
		}
		if j.Name == "" || len(j.Name) > 10 {
			log.Println("Player sent invalid name")
			return
		}
		if strings.HasPrefix(strings.ToUpper(j.Name), "PLAYER") {
			log.Println("Player is a reserved prefix")
			return
		}
		for _, t := range game.Tanks {
			if strings.ToUpper(t.Name) == strings.ToUpper(j.Name) {
				log.Println("Player tried to name themself the same as another")
				return
			}
		}
		fmt.Println(tank.Name, "is now", j.Name)
		tank.Name = j.Name
		j.ID = tank.ID
		j.Color = f32.Vec3(tank.Color)
		Players.Each(func(id tanklets.PlayerID, p *net.UDPAddr) {
			if tank.ID == id {
				j.You = 1
			} else {
				j.You = 0
			}
			game.Network.Send(j, p)
		})
		return
	}

	Players.Put(playerId, addr)

	if game.State != tanklets.StateWaiting {
		Players.Each(func(id tanklets.PlayerID, p *net.UDPAddr) {
			if id == playerId {
				return
			}
			// tell this player where all the existing players are
			thisTank := game.Tanks[id]
			game.Network.Send(pkt.Join{id, 0, f32.Vec3(thisTank.Color), thisTank.Name}, addr)
			game.Network.Send(thisTank.Location(), addr)
		})
		return
	}

	tank = game.NewTank(playerId, pkt.GetColor(game.CursorColor.Next()))
	tank.LastPkt = time.Now()
	tank.SetPosition(cp.Vector{
		X: 10 + float64(rand.Intn(int(game.Width)-20)),
		Y: 10 + float64(rand.Intn(int(game.Height)-20)),
	})

	// tell this player their ID
	join := pkt.Join{tank.ID, 1, f32.Vec3(tank.Color), tank.Name}
	game.Network.Send(join, addr)
	loc := tank.Location()
	// tell this player where they are
	game.Network.Send(loc, addr)
	join.You = 0
	Players.Each(func(id tanklets.PlayerID, p *net.UDPAddr) {
		if id == tank.ID {
			return
		}
		// tell all players about this player
		game.Network.Send(join, p)
		game.Network.Send(loc, p)
		// tell this player where all the existing players are
		thisTank := game.Tanks[id]
		game.Network.Send(pkt.Join{id, 0, f32.Vec3(thisTank.Color), thisTank.Name}, addr)
		game.Network.Send(thisTank.Location(), addr)
	})
	// Tell this player about the level
	for _, box := range game.Boxes {
		game.Network.Send(box.Location(), addr)
	}
	game.Tanks[tank.ID] = tank
	fmt.Println("tank", tank.ID, "joined")
	HasHadPlayersConnect = true
}

func disconnect(packet tanklets.Packet, game *Game) {
	addr := packet.Addr

	playerID, _ := Players.Lookup(addr.String())
	player := Players.Get(playerID)
	if player == nil {
		// this is normal, we spam disconnect when leaving to ensure the server gets it
		return
	}

	Players.Delete(playerID)
	if game.Tanks[playerID] != nil {
		game.Tanks[playerID].Destroyed = true
	}

	// tell others they left & destroyed
	Players.SendAll(game.Network, pkt.Disconnect{ID: uint16(playerID)}, pkt.Damage{ID: playerID, Killer: playerID})
}

func move(packet tanklets.Packet, game *Game) {
	if game.State < tanklets.StatePlaying {
		return
	}

	m := pkt.Move{}
	if _, err := m.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}
	addr := packet.Addr

	id, _ := Players.Lookup(addr.String())
	tank := game.Tanks[id]
	if tank == nil {
		log.Println("Player not found", addr.String(), id)
		return
	}
	if tank.Destroyed {
		return
	}

	tank.NextMove.Turn = m.Turn
	tank.NextMove.Throttle = m.Throttle
	tank.NextMove.TurretAngle = m.TurretAngle

	tank.ControlBody.SetAngularVelocity(float64(m.Turn) * tanklets.TurnSpeed)
	tank.ControlBody.SetVelocityVector(tank.Body.Rotation().Rotate(cp.Vector{Y: float64(m.Throttle) * tanklets.MaxSpeed}))

	tank.Turret.SetPosition(tank.Body.Position())
	tank.Turret.SetAngle(m.TurretAngle)
}

func ready(packet tanklets.Packet, game *Game) {
	id, _ := Players.Lookup(packet.Addr.String())
	tank := game.Tanks[id]
	if tank != nil {
		fmt.Println("Got a ready from", tank.ID)
		tank.Ready = true
	}
}

func shoot(packet tanklets.Packet, game *Game) {
	addr := packet.Addr
	id, _ := Players.Lookup(addr.String())
	player := Players.Get(id)
	if player == nil {
		log.Println("Player not found", addr.String(), id)
		return
	}
	tank := game.Tanks[id]

	if tank.Destroyed {
		return
	}

	if time.Now().Sub(tank.LastShot) < tanklets.ShotCooldown {
		return
	}
	tank.LastShot = time.Now()

	shot := pkt.Shoot{}
	_, err := shot.Serialize(packet.Bytes)
	if err != nil {
		log.Println("Shoot packet is bad", err)
		return
	}

	bullet := game.NewBullet(tank, tanklets.BulletID(game.CursorBullet.Next()))

	pos := cp.Vector{X: tanklets.TankHeight / 2.0}
	pos = pos.Rotate(cp.Vector{X: math.Cos(shot.Angle), Y: math.Sin(shot.Angle)})
	bullet.Body.SetPosition(pos.Add(tank.Turret.Position()))
	bullet.Body.SetAngle(shot.Angle)
	bullet.Body.SetVelocityVector(bullet.Body.Rotation().Rotate(cp.Vector{tanklets.BulletSpeed, 0}))
	//bullet.Shape.SetFilter(cp.NewShapeFilter(uint(player.ID), cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	Players.SendAll(game.Network, bullet.Location())
}

func ping(packet tanklets.Packet, game *Game) {
	addr := packet.Addr
	id, _ := Players.Lookup(addr.String())
	player := Players.Get(id)
	if player == nil {
		log.Println("Player not found", addr.String(), id)
		return
	}
	tank := game.Tanks[id]
	if tank == nil {
		log.Println("Timing issue?")
		return
	}
	tank.LastPkt = time.Now()
}
