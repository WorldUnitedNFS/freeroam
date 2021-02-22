package carstate

import (
	"github.com/WorldUnitedNFS/freeroam/math"
)

type AirPacket struct {
	PacketStruct

	Yaw   float64
	Pitch float64
	Roll  float64
}

func NewAirPacket(simTime uint16) AirPacket {
	pkt := AirPacket{}
	pkt.simTime = simTime
	return pkt
}

func (g AirPacket) SimTime() uint16 {
	return g.simTime
}

func (g AirPacket) OnGround() bool {
	return false
}

func (g AirPacket) Coordinates() math.Vector3D {
	return math.Vector3D{
		X: g.posX,
		Y: g.posY,
		Z: g.posZ,
	}
}

func (g AirPacket) Rotation() float64 {
	return math.RadToDeg(g.Yaw)
}

func (g AirPacket) LinearVelocity() math.Vector3D {
	return math.Vector3D{
		X: g.linVelX,
		Y: g.linVelY,
		Z: g.linVelZ,
	}
}

func (g AirPacket) AngularVelocity() math.Vector3D {
	return math.Vector3D{
		X: g.angVelX,
		Y: g.angVelY,
		Z: g.angVelZ,
	}
}

func (g *AirPacket) Decode(reader *PacketReader) error {
	yaw, err := reader.DecodeFloat(9, 0xE0, 0.5, 0.007853982, -3.1415927)
	if err != nil {
		return err
	}
	pitch, err := reader.DecodeFloat(8, 0x70, 0.5, 0.007853982, -1.5707964)
	if err != nil {
		return err
	}
	roll, err := reader.DecodeFloat(9, 0xE0, 0.5, 0.007853982, -3.1415927)

	if err != nil {
		return err
	}

	posY, err := reader.DecodeFloat(17, 62144, 0.5, 0.15000001, -15000)
	if err != nil {
		return err
	}

	posZ, err := reader.DecodeFloat(11, 96, 0.5, 0.639999999, -512)

	if err != nil {
		return err
	}

	posX, err := reader.DecodeFloat(17, 62144, 0.5, 0.15000001, -15000)

	if err != nil {
		return err
	}

	linVelY, err := reader.DecodeFloat(9, 0x7B, 0.5, 0.30829942, -138.8889)

	if err != nil {
		return err
	}

	linVelZ, err := reader.DecodeFloat(9, 0x7B, 0.5, 0.30829942, -138.8889)

	if err != nil {
		return err
	}

	linVelX, err := reader.DecodeFloat(9, 0x7B, 0.5, 0.30829942, -138.8889)

	if err != nil {
		return err
	}

	angVelY, err := reader.DecodeFloat(8, 0xD3, 0.5, 0.12524623, -18.849556)

	if err != nil {
		return err
	}

	angVelZ, err := reader.DecodeFloat(8, 0xD3, 0.5, 0.12524623, -18.849556)

	if err != nil {
		return err
	}

	angVelX, err := reader.DecodeFloat(8, 0xD3, 0.5, 0.12524623, -18.849556)

	if err != nil {
		return err
	}

	g.posX = posX
	g.posY = -posY
	g.posZ = posZ
	g.Yaw = yaw
	g.Roll = roll
	g.Pitch = pitch
	g.angVelX = angVelX
	g.angVelY = -angVelY
	g.angVelZ = angVelZ
	g.linVelX = linVelX
	g.linVelY = -linVelY
	g.linVelZ = linVelZ

	return nil
}
