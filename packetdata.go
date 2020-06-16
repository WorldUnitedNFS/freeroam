package freeroam

import "bytes"

type PacketReader struct {
	BitReader *BitReader
}

func NewPacketReader(packet []byte) *PacketReader {
	return &PacketReader{
		BitReader: NewReader(bytes.NewReader(packet)),
	}
}

func (packetReader *PacketReader) ReadInitialBits() error {
	_, err := packetReader.BitReader.ReadBits(28)
	return err
}

func (packetReader *PacketReader) DecodeYCoordinate() (float64, error) {
	yBits, err := packetReader.BitReader.ReadBits(17)

	if err != nil {
		return 0.0, err
	}

	if yBits >= 62144 {
		hasDataLeft := packetReader.BitReader.HasDataLeft()
		var iHasDataLeft uint64
		if hasDataLeft {
			iHasDataLeft = 1
		}
		yBits = iHasDataLeft + 2*yBits - 62144
	}

	yBitsFloat := float64(yBits)
	return -((yBitsFloat+0.5)*0.039999999 - 5000.0), nil
}

func (packetReader *PacketReader) DecodeZCoordinate() (float64, error) {
	zBits, err := packetReader.BitReader.ReadBits(11)

	if err != nil {
		return 0.0, err
	}

	if zBits >= 96 {
		hasDataLeft := packetReader.BitReader.HasDataLeft()
		var iHasDataLeft uint64
		if hasDataLeft {
			iHasDataLeft = 1
		}
		zBits = iHasDataLeft + 2*zBits - 96
	}

	zBitsFloat := float64(zBits)
	return (zBitsFloat+0.5)*0.12774999 - 112.0, nil
}

func (packetReader *PacketReader) DecodeXCoordinate() (float64, error) {
	xBits, err := packetReader.BitReader.ReadBits(17)

	if err != nil {
		return 0.0, err
	}

	if xBits >= 62144 {
		hasDataLeft := packetReader.BitReader.HasDataLeft()
		var iHasDataLeft uint64
		if hasDataLeft {
			iHasDataLeft = 1
		}
		xBits = iHasDataLeft + 2*xBits - 62144
	}

	xBitsFloat := float64(xBits)
	return (xBitsFloat + 0.5) * 0.059999999, nil
}
