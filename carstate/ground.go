package carstate

import (
	"github.com/WorldUnitedNFS/freeroam/math"
	"github.com/westphae/quaternion"
)

type GroundPacket struct {
	PacketStruct

	FrontWheelsDirection  float64
	RearWheelsDirection   float64
	ActiveEffectFlags     uint32
	OrientationQuaternion quaternion.Quaternion
	RollRadians           float64
}

func NewGroundPacket(simTime uint16) GroundPacket {
	pkt := GroundPacket{}
	pkt.simTime = simTime
	return pkt
}

func (g GroundPacket) SimTime() uint16 {
	return g.simTime
}

func (g GroundPacket) OnGround() bool {
	return true
}

func (g GroundPacket) Coordinates() math.Vector3D {
	return math.Vector3D{
		X: g.posX,
		Y: g.posY,
		Z: g.posZ,
	}
}

func (g GroundPacket) Rotation() float64 {
	return math.RadToDeg(-g.RollRadians)
}

func (g GroundPacket) LinearVelocity() math.Vector3D {
	return math.Vector3D{
		X: g.linVelX,
		Y: g.linVelY,
		Z: g.linVelZ,
	}
}

func (g GroundPacket) AngularVelocity() math.Vector3D {
	return math.Vector3D{
		X: g.angVelX,
		Y: g.angVelY,
		Z: g.angVelZ,
	}
}

func (g GroundPacket) Orientation() quaternion.Quaternion {
	return g.OrientationQuaternion
}

func (g *GroundPacket) Decode(reader *PacketReader) error {
	posY, err := reader.DecodeFloat(17, 62144, 0.5, 0.039999999, -5000)
	if err != nil {
		return err
	}

	posZ, err := reader.DecodeFloat(11, 96, 0.5, 0.12774999, -112)

	if err != nil {
		return err
	}

	posX, err := reader.DecodeFloat(17, 62144, 0.5, 0.059999999, 0)

	if err != nil {
		return err
	}

	// 007ECB61
	linVelY, err := reader.DecodeFloat(14, 0x31E0, 0.5, 0.016666668, -166.66667)

	if err != nil {
		return err
	}

	// 007ECB61
	linVelZ, err := reader.DecodeFloat(10, 0x350, 0.5, 0.1388889, -83.333336)

	if err != nil {
		return err
	}

	// 007ECB61
	linVelX, err := reader.DecodeFloat(14, 0x31E0, 0.5, 0.016666668, -166.66667)

	if err != nil {
		return err
	}

	// 007ECD41
	orientationY, err := reader.DecodeFloat(8, 1, 0.5, 0.0039138943, -1.0)

	if err != nil {
		return err
	}

	// 007ECE61
	orientationZ, err := reader.DecodeFloat(9, 0xE0, 0.5, 0.0024999999, -1.0)

	if err != nil {
		return err
	}

	// 007ECD41
	orientationX, err := reader.DecodeFloat(8, 1, 0.5, 0.0039138943, -1.0)

	if err != nil {
		return err
	}

	// 007ECE61
	orientationW, err := reader.DecodeFloat(9, 0xE0, 0.5, 0.0024999999, -1.0)

	if err != nil {
		return err
	}

	// 007ECAA1
	angVelY, err := reader.DecodeFloat(8, 0xD3, 0.5, 0.12524623, -18.849556)

	if err != nil {
		return err
	}

	// 007ECAA1
	angVelZ, err := reader.DecodeFloat(8, 0xD3, 0.5, 0.12524623, -18.849556)

	if err != nil {
		return err
	}

	// 007ECAA1
	angVelX, err := reader.DecodeFloat(8, 0xD3, 0.5, 0.12524623, -18.849556)

	if err != nil {
		return err
	}

	// 007ECDA1
	frontWheelsDirection, err := reader.DecodeFloat(6, 0x1B, 0.5, 0.021980198, -1.11)

	if err != nil {
		return err
	}

	// 007ECDA1
	rearWheelsDirection, err := reader.DecodeFloat(6, 0x1B, 0.5, 0.021980198, -1.11)

	if err != nil {
		return err
	}

	lightFlags, err := reader.BitReader.ReadBits(13)

	if err != nil {
		return err
	}

	g.posX = posX
	g.posY = -posY
	g.posZ = posZ
	g.angVelX = angVelX
	g.angVelY = -angVelY
	g.angVelZ = angVelZ
	g.linVelX = linVelX
	g.linVelY = -linVelY
	g.linVelZ = linVelZ
	g.OrientationQuaternion = quaternion.Quaternion{
		X: orientationX,
		Y: orientationY,
		Z: orientationZ,
		W: orientationW,
	}
	g.ActiveEffectFlags = lightFlags
	g.FrontWheelsDirection = frontWheelsDirection
	g.RearWheelsDirection = rearWheelsDirection

	_, _, roll := g.OrientationQuaternion.Euler()
	g.RollRadians = roll

	return nil
}
