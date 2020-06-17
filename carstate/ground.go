package carstate

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

	return nil
}
