package tanklets

import (
	"log"
	"net"
	"math"

	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/binserializer"
)

// message sent to clients: update location information (58 bytes)
type Location struct {
	ID                     PlayerID // 2 bytes
	X, Y                   float64  // 16 bytes
	Vx, Vy                 float64  // 16
	Angle, AngularVelocity float64  // 16

	Turret float64 // 8
}

func (l *Location) Handle(addr *net.UDPAddr) {
	if IsServer {
		log.Println("I shouldn't have gotten this")
		return
	}

	player := Tanks[l.ID]
	if player == nil {
		log.Println("Client", Me, "-- Player with ID", l.ID, "not found")
		return
	}
	pos := player.Body.Position()
	newPos := cp.Vector{l.X, l.Y}

	if pos.Distance(newPos) > 1 {
		player.Body.SetPosition(newPos)
	} else {
		player.Body.SetPosition(pos.Lerp(newPos, 0.5))
	}

	angle := player.ControlBody.Angle()
	if math.Abs(angle-l.Angle) > 5 {
		player.ControlBody.SetAngle(l.Angle)
	} else {
		player.ControlBody.SetAngle(cp.Lerp(l.Angle, angle, 0.5))
	}

	player.Turret.SetPosition(player.Body.Position())
	player.ControlBody.SetVelocity(l.Vx, l.Vy)
	player.ControlBody.SetAngularVelocity(l.AngularVelocity)
	player.Turret.Body.SetAngle(l.Turret)
}

func (l *Location) MarshalBinary() ([]byte, error) {
	buf := binserializer.NewBuffer(59)
	buf.WriteByte(LOCATION)
	buf.WriteUint16(uint16(l.ID))
	buf.WriteFloat64(l.X)
	buf.WriteFloat64(l.Y)
	buf.WriteFloat64(l.Vx)
	buf.WriteFloat64(l.Vy)
	buf.WriteFloat64(l.Angle)
	buf.WriteFloat64(l.AngularVelocity)
	buf.WriteFloat64(l.Turret)
	return buf.Bytes()
}

func (l *Location) UnmarshalBinary(b []byte) error {
	buf := binserializer.NewBufferFromBytes(b)
	_ = buf.GetByte()
	l.ID = PlayerID(buf.GetUint16())
	l.X = buf.GetFloat64()
	l.Y = buf.GetFloat64()
	l.Vx = buf.GetFloat64()
	l.Vy = buf.GetFloat64()
	l.Angle = buf.GetFloat64()
	l.AngularVelocity = buf.GetFloat64()
	l.Turret = buf.GetFloat64()
	return buf.Error()
}
