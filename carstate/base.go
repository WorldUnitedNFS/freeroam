package carstate

import "github.com/WorldUnitedNFS/freeroam/math"

type Packet interface {
	SimTime() uint16
	OnGround() bool
	Coordinates() math.Vector3D
	Rotation() float64
	LinearVelocity() math.Vector3D
	AngularVelocity() math.Vector3D

	Decode(reader *PacketReader) error
}

type PacketStruct struct {
	simTime uint16
	posX    float64
	posY    float64
	posZ    float64
	linVelX float64
	linVelY float64
	linVelZ float64
	angVelX float64
	angVelY float64
	angVelZ float64
}
