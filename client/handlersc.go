package client

import (
	"github.com/jakecoffman/tanklets"
	"fmt"
	"log"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets/pkt"
)

type packetHandler func(packet tanklets.Packet, game *tanklets.Game, network *Client)

var handlers [pkt.PacketMax]packetHandler

func init() {
	for i := 0; i < pkt.PacketMax; i++ {
		handlers[i] = noop
	}

	handlers[pkt.PacketInit] = initial
	handlers[pkt.PacketJoin] = join
	handlers[pkt.PacketLocation] = location
	handlers[pkt.PacketState] = state
	handlers[pkt.PacketDisconnect] = disconnect
	handlers[pkt.PacketBoxLocation] = boxlocation
	handlers[pkt.PacketDamage] = damage
	handlers[pkt.PacketShoot] = shoot
}

func ProcessNetwork(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	handlers[packet.Bytes[0]](packet, game, network)
}

func noop(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	log.Println("Unhandled client packet", packet.Bytes[0])
}

func initial(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	initial := pkt.Initial{}
	if _, err := initial.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("I am connected!")
	Me = tanklets.PlayerID(initial.ID)
	network.IsConnected = true
	network.IsConnecting = false
}

func join(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	j := pkt.Join{}
	if _, err := j.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	tank := game.Tanks[j.ID]
	if tank != nil {
		// player already joined, this is informational
		tank.Color = mgl32.Vec3(j.Color)
		fmt.Println(tank.Name, "is now", j.Name)
		tank.Name = j.Name
		return
	}

	fmt.Println("Player joined")
	tank = game.NewTank(j.ID, mgl32.Vec3(j.Color))
	if j.You > 0 {
		fmt.Println("Oh, it's me!")
		Me = tank.ID
	}
	game.Tanks[tank.ID] = tank
}

var lastLocationSeq uint64

func location(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	l := pkt.Location{}
	if _, err := l.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	if l.Sequence < lastLocationSeq {
		fmt.Println("out of order packet")
		return
	}
	lastLocationSeq = l.Sequence

	player := game.Tanks[tanklets.PlayerID(l.ID)]
	if player == nil {
		log.Println("Client", Me, "-- Tank with ID", l.ID, "not found")
		return
	}
	pos := player.Position()
	newPos := cp.Vector{float64(l.X), float64(l.Y)}

	diff := newPos.Sub(pos)
	distance := diff.Length()

	// https://gafferongames.com/post/networked_physics_2004/
	if distance > 6 {
		player.SetPosition(newPos)
	} else {
		player.SetPosition(pos.Add(diff.Mult(0.1)))
	}
	player.Turret.SetPosition(player.Body.Position())

	player.SetAngle(float64(l.Angle))
	player.ControlBody.SetAngle(player.Angle())

	player.SetVelocity(float64(l.Vx), float64(l.Vy))
	player.ControlBody.SetVelocityVector(player.Velocity())
	player.SetAngularVelocity(float64(l.AngularVelocity))
	player.ControlBody.SetAngularVelocity(player.AngularVelocity())
	player.Turret.Body.SetAngle(float64(l.Turret))
}

var lastBoxSeq uint64

func boxlocation(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	l := pkt.BoxLocation{}
	if _, err := l.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	if l.Sequence < lastBoxSeq {
		return
	}
	lastBoxSeq = l.Sequence

	box := game.Boxes[tanklets.BoxID(l.ID)]
	if box == nil {
		box = game.NewBox(tanklets.BoxID(l.ID))
	}
	pos := box.Position()
	newPos := cp.Vector{float64(l.X), float64(l.Y)}

	diff := newPos.Sub(pos)
	distance := diff.Length()

	if distance > 4 {
		box.SetPosition(newPos)
	} else {
		box.SetPosition(pos.Add(diff.Mult(0.1)))
	}

	box.SetAngle(float64(l.Angle))
}

func damage(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	d := pkt.Damage{}
	if _, err := d.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	tank := game.Tanks[tanklets.PlayerID(d.ID)]
	if tank == nil {
		log.Println("Tank", d.ID, "not found")
		return
	}
	tank.Destroyed = true
}

func disconnect(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	d := pkt.Disconnect{}
	if _, err := d.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Client", Me, "-- Player", d.ID, "Has disonnceted")
}

func shoot(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	s := pkt.Shoot{}
	if _, err := s.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	firedBy := game.Tanks[tanklets.PlayerID(s.PlayerID)]
	bullet := game.Bullets[tanklets.BulletID(s.BulletID)]
	if bullet == nil {
		bullet = game.NewBullet(firedBy, tanklets.BulletID(s.BulletID))
		game.Bullets[tanklets.BulletID(s.BulletID)] = bullet
	}

	bullet.Bounce = int(s.Bounce)

	if bullet.Bounce > 1 {
		bullet.Destroy(true)
		return
	}

	bullet.Body.SetPosition(cp.Vector{s.X, s.Y})
	bullet.Body.SetAngle(s.Angle)
	bullet.Body.SetVelocity(s.Vx, s.Vy)
}

func state(packet tanklets.Packet, game *tanklets.Game, network *Client) {
	s := pkt.State{}
	if _, err := s.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	switch s.State {
	case tanklets.StateStartCountdown:
		// start the countdown
		game.StartTime = time.Now()
		// this might be a game reset, so clean up some
		for _, t := range game.Tanks {
			t.Destroyed = false
			t.SetVelocityVector(cp.Vector{})
			t.SetAngularVelocity(0)
			t.SetAngle(0)
			t.LastMove = pkt.Move{}
			t.NextMove = pkt.Move{}
		}
		game.WinningPlayer = nil
	case tanklets.StateWinCountdown:
		tank := game.Tanks[s.ID]
		game.WinningPlayer = tank
		tank.Score++
	case tanklets.StateFailCountdown:
		if game.WinningPlayer != nil {
			game.WinningPlayer.Score--
		}
	}

	game.State = int(s.State)
}
