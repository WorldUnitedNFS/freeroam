package main

import (
	"fmt"
	"gitlab.com/sparkserver/freeroam"
)

func main() {
	testPacket := []byte{
		0x00, 0x00, 0x98, 0x0a, 0x1e, 0xb1, 0x83, 0x10,
		0x6f, 0xf3, 0x88, 0x4b, 0x13, 0x88, 0x40, 0xb5,
		0xa7, 0xf7, 0x69, 0x2d, 0x2d, 0x2d, 0x35, 0xf8,
		0x00, 0x3f,
	}

	testPacketReader := freeroam.NewPacketReader(testPacket)

	// skip initial bits
	err := testPacketReader.ReadInitialBits()

	if err != nil {
		panic(err)
	}

	y, err := testPacketReader.DecodeYCoordinate()

	if err != nil {
		panic(err)
	}

	z, err := testPacketReader.DecodeZCoordinate()

	if err != nil {
		panic(err)
	}

	x, err := testPacketReader.DecodeXCoordinate()

	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println(x)
	fmt.Println(y)
	fmt.Println(z)
}
