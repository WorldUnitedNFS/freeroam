package carstate

import "fmt"

type GroundPacket struct {
	simTime uint16
	posX    float64
	posY    float64
	posZ    float64
}

func NewGroundPacket(simTime uint16) GroundPacket {
	return GroundPacket{
		simTime: simTime,
	}
}

func (g GroundPacket) SimTime() uint16 {
	return g.simTime
}

func (g GroundPacket) OnGround() bool {
	return true
}

func (g GroundPacket) XPos() float64 {
	return g.posX
}

func (g GroundPacket) YPos() float64 {
	return g.posY
}

func (g GroundPacket) ZPos() float64 {
	return g.posZ
}

func (g *GroundPacket) Decode(reader *PacketReader) error {
	posY, err := reader.DecodeFloat(17, 62144, 0.5, 0.039999999, -5000)
	if err != nil {
		return err
	}
	posY *= -1

	posZ, err := reader.DecodeFloat(11, 96, 0.5, 0.12774999, -112)

	if err != nil {
		return err
	}

	posX, err := reader.DecodeFloat(17, 62144, 0.5, 0.059999999, 0)

	if err != nil {
		return err
	}

	g.posX = posX
	g.posY = posY
	g.posZ = posZ

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

	fmt.Println("decoded ground packet")
	fmt.Printf("\tCoordinates          = (%f, %f, %f)\n", posX, posY, posZ)
	fmt.Printf("\tLinear Velocity      = (%f, %f, %f)\n", linVelX, linVelY, linVelZ)
	fmt.Printf("\tOrientation          = (%f, %f, %f, %f)\n", orientationX, orientationY, orientationZ, orientationW)
	fmt.Printf("\tAngular Velocity     = (%f, %f, %f)\n", angVelX, angVelY, angVelZ)
	fmt.Printf("\tFrontWheelsDirection = %f\n", frontWheelsDirection)
	fmt.Printf("\tRearWheelsDirection  = %f\n", rearWheelsDirection)
	fmt.Printf("\tLightFlags           = %d\n", lightFlags)
	return nil
}
