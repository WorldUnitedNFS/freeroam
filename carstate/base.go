package carstate

type BasePacket interface {
	SimTime() uint16
	OnGround() bool
	XPos() float64
	YPos() float64
	ZPos() float64

	Decode(reader *PacketReader) error
}
