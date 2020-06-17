package carstate

import (
	"bytes"
	"fmt"
)

type PacketReader struct {
	BitReader *BitReader
}

func NewPacketReader(packet []byte) *PacketReader {
	return &PacketReader{
		BitReader: NewReader(bytes.NewReader(packet)),
	}
}

func (packetReader *PacketReader) Decode() (BasePacket, error) {
	// Read header, starting with time value
	simTimeResult, simTimeErr := packetReader.BitReader.ReadBits(16)
	if simTimeErr != nil {
		return nil, simTimeErr
	}
	simTime := uint16(simTimeResult)

	_, err := packetReader.BitReader.ReadBits(2)

	if err != nil {
		return nil, err
	}

	_, err = packetReader.BitReader.ReadBit()

	if err != nil {
		return nil, err
	}

	_, err = packetReader.BitReader.ReadBit()

	if err != nil {
		return nil, err
	}

	ground, err := packetReader.BitReader.ReadBit()

	if err != nil {
		return nil, err
	}

	_, err = packetReader.BitReader.ReadBit()

	if err != nil {
		return nil, err
	}

	_, err = packetReader.BitReader.ReadBits(6)

	if err != nil {
		return nil, err
	}

	if ground {
		groundPkt := NewGroundPacket(simTime)
		err = groundPkt.Decode(packetReader)

		if err != nil {
			return nil, err
		}

		return &groundPkt, nil
	}

	return nil, fmt.Errorf("cannot handle packet")
}

// DecodeFloat decodes a compressed floating point value from a packet.
func (packetReader *PacketReader) DecodeFloat(numBits int, maxValue uint64, addValue1 float64, multiplyValue1 float64, addValue2 float64) (float64, error) {
	rawBits, err := packetReader.BitReader.ReadBits(numBits)

	if err != nil {
		return 0, err
	}

	if rawBits >= maxValue {
		hasDataLeft := packetReader.BitReader.HasDataLeft()
		var iHasDataLeft uint64
		if hasDataLeft {
			iHasDataLeft = 1
		}
		rawBits = iHasDataLeft + 2*rawBits - maxValue
	}

	rawBitsFloat := float64(rawBits)
	return (rawBitsFloat+addValue1)*multiplyValue1 + addValue2, nil
}
