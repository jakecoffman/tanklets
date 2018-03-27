package pkt

const (
	PacketInit = iota
	PacketJoin
	PacketReady
	PacketState
	PacketDisconnect
	PacketMove
	PacketShoot
	PacketBulletUpdate
	PacketLocation
	PacketBoxLocation
	PacketDamage
	PacketPing
	PacketAim
	PacketMax
)

type PlayerID uint16
type BoxID uint16
type BulletID uint64
