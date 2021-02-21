package main

import (
	"fmt"
	"github.com/WorldUnitedNFS/freeroam/binary"
	"github.com/WorldUnitedNFS/freeroam/carstate"
)

func main() {
	testPacket := []byte{
		0x2E, 0xA6, 0x90, 0x0E, 0x62, 0x6F, 0x45, 0xCB,
		0xFA, 0x27, 0xA9, 0x7E, 0x6E, 0x57, 0x0F, 0x4B,
		0x93, 0x2B, 0x2D, 0x2B, 0x36, 0x68, 0x18, 0x7F,
	}

	bitstream := binary.NewBitstream(testPacket)
	val, err := bitstream.ReadBits(16)
	if err != nil {
		panic(err)
	}
	fmt.Println(val)

	reader := carstate.NewPacketReader(testPacket)
	pkt, err := reader.Decode()

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", pkt)
}
