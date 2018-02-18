package pkt

const (
	PacketInit = iota
	PacketJoin
	PacketReady
	PacketState
	PacketDisconnect
	PacketMove
	PacketShoot
	PacketLocation
	PacketBoxLocation
	PacketDamage
	PacketPing
	PacketMax
)

type PlayerID uint16
type BoxID uint16
type BulletID uint64
