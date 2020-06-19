// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"encoding/binary"
	"gitlab.com/sparkserver/freeroam/carstate"
	"gitlab.com/sparkserver/freeroam/math"
)

// WriteSubpacket writes a subpacket with specified type and payload to a bytes.Buffer
func WriteSubpacket(buf *bytes.Buffer, typ uint8, data []byte) {
	if len(data) > 255 {
		panic("WriteSubpacket: subpacket length exceeds 255")
	}
	buf.WriteByte(typ)
	buf.WriteByte(uint8(len(data)))
	buf.Write(data)
}

type CarPosPacket struct {
	time   uint16
	packet []byte
	pos    math.Vector3D
}

// Valid returns true if CarPosPacket contains valid packet data.
func (p *CarPosPacket) Valid() bool {
	return p.packet != nil
}

// Pos returns the car position as a Vector2D.
func (p *CarPosPacket) Pos() math.Vector3D {
	return p.pos
}

// Packet returns the packet data with the packet time replaced by the argument.
func (p *CarPosPacket) Packet(time uint16) []byte {
	binary.BigEndian.PutUint16(p.packet, time)
	return p.packet
}

// Update updates CarPosPacket with the specified byte slice.
// The supplied slice shouldn't be modified after calling this method.
func (p *CarPosPacket) Update(packet []byte) {
	p.time = binary.BigEndian.Uint16(packet[0:2])
	p.packet = packet
	reader := carstate.NewPacketReader(packet)
	decodedPacket, err := reader.Decode()

	if err != nil {
		panic(err)
	}

	coords := decodedPacket.Coordinates()
	p.pos = coords
}
