package client

import (
	"github.com/jakecoffman/tanklets"
	"fmt"
	"log"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type packetHandler func(packet tanklets.Packet, game *tanklets.Game)

var handlers [tanklets.PacketMax]packetHandler

func init() {
	for i := 0; i < tanklets.PacketMax; i++ {
		handlers[i] = noop
	}

	handlers[tanklets.PacketInit] = initial
	handlers[tanklets.PacketJoin] = join
	handlers[tanklets.PacketLocation] = location
	handlers[tanklets.PacketState] = state
	handlers[tanklets.PacketDisconnect] = disconnect
	handlers[tanklets.PacketBoxLocation] = boxlocation
	handlers[tanklets.PacketDamage] = damage
	handlers[tanklets.PacketShoot] = shoot
}

func ProcessNetwork(packet tanklets.Packet, game *tanklets.Game) {
	handlers[packet.Bytes[0]](packet, game)
}

func noop(packet tanklets.Packet, _ *tanklets.Game) {
	log.Println("Unhandled client packet", packet.Bytes[0])
}

func initial(packet tanklets.Packet, _ *tanklets.Game) {
	initial := tanklets.Initial{}
	if _, err := initial.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("I am connected!")
	tanklets.Me = initial.ID
	tanklets.ClientIsConnected = true
	tanklets.ClientIsConnecting = false
}

func join(packet tanklets.Packet, game *tanklets.Game) {
	j := tanklets.Join{}
	if _, err := j.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Player joined")
	tank := game.NewTank(j.ID, mgl32.Vec3(j.Color))
	if j.You > 0 {
		fmt.Println("Oh, it's me!")
		tanklets.Me = tank.ID
		//Player = player
	}
	game.Tanks[tank.ID] = tank
}

func location(packet tanklets.Packet, game *tanklets.Game) {
	l := tanklets.Location{}
	if _, err := l.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	player := game.Tanks[l.ID]
	if player == nil {
		log.Println("Client", tanklets.Me, "-- Player with ID", l.ID, "not found")
		return
	}
	pos := player.Position()
	newPos := cp.Vector{float64(l.X), float64(l.Y)}

	diff := newPos.Sub(pos)
	distance := diff.Length()

	// https://gafferongames.com/post/networked_physics_2004/
	if distance > 4 {
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

func boxlocation(packet tanklets.Packet, game *tanklets.Game) {
	l := tanklets.BoxLocation{}
	if _, err := l.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	box := game.Boxes[l.ID]
	if box == nil {
		box = game.NewBox(l.ID)
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

func damage(packet tanklets.Packet, game *tanklets.Game) {
	d := tanklets.Damage{}
	if _, err := d.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	tank := game.Tanks[d.ID]
	if tank == nil {
		log.Println("Tank", d.ID, "not found")
		return
	}
	tank.Destroyed = true

	if d.ID == tanklets.Me {
		game.State = tanklets.GameStateDead
	}
}

func disconnect(packet tanklets.Packet, game *tanklets.Game) {
	d := tanklets.Disconnect{}
	if _, err := d.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Client", tanklets.Me, "-- Player", d.ID, "Has disonnceted")
}

func shoot(packet tanklets.Packet, game *tanklets.Game) {
	s := tanklets.Shoot{}
	if _, err := s.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	firedBy := game.Tanks[s.PlayerID]
	bullet := game.Bullets[s.BulletID]
	if bullet == nil {
		bullet = game.NewBullet(firedBy, s.BulletID)
		game.Bullets[s.BulletID] = bullet
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

func state(packet tanklets.Packet, game *tanklets.Game) {
	s := tanklets.State{}
	if _, err := s.Serialize(packet.Bytes); err != nil {
		log.Println(err)
		return
	}

	game.State = int(s.State)
}
