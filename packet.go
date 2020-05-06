// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package freeroam

import (
	"bytes"
	"encoding/binary"
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
	pos    Vector
}

// Valid returns true if CarPosPacket contains valid packet data.
func (p *CarPosPacket) Valid() bool {
	return p.packet != nil
}

// Pos returns the car position as a Vector.
func (p *CarPosPacket) Pos() Vector {
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
	flying := (packet[2] >> 3) & 1
	if flying == 1 {
		p.pos.X = p.getX()
		p.pos.Y = p.getY()
	}
}

func clone(a []byte) []byte {
	out := make([]byte, len(a))
	copy(out, a)
	return out
}

func (p *CarPosPacket) getY() float64 {
	out := clone(p.packet[3:6])
	var shift uint
	if out[0] >= 7 {
		shift = 2
	} else {
		shift = 3
	}
	nv := float64(binary.BigEndian.Uint32([]byte{0x00, out[0], out[1], out[2]})>>shift) / 25
	var f float64
	if out[0] >= 7 {
		f = 5000 - nv + 2378.6
	} else {
		f = 5000 - nv
	}
	return f
}

func (p *CarPosPacket) getX() float64 {
	out := clone(p.packet[7:10])
	var shift uint
	out[0] = out[0] & 0x3f
	if p.packet[7]&32 > 0 {
		shift = 5
	} else {
		shift = 6
	}
	nv := float64(binary.BigEndian.Uint32([]byte{0x00, out[0], out[1], out[2]}) >> shift)
	i := nv / 8.332
	if shift == 5 {
		i -= 3730
	} else if p.packet[3] >= 8 || p.packet[3] == 7 && p.packet[4] >= 128 {
		//
	} else {
		i += 4135
	}
	return i
}
