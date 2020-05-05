// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package internal

import (
	"encoding/binary"
	"fmt"
)

type CarPosPacket struct {
	time   uint16
	packet []byte
	pos    Vector2D
	lastY  float64
	lastX  float64
}

func (p *CarPosPacket) Valid() bool {
	return p.packet != nil
}

func (p *CarPosPacket) Pos() Vector2D {
	return p.pos
}

func (p *CarPosPacket) Packet(timei uint16) []byte {
	time := make([]byte, 2)
	binary.BigEndian.PutUint16(time, timei)
	p.packet[0] = time[0]
	p.packet[1] = time[1]
	return p.packet
}

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
	if nv != p.lastY {
		printBinary(p.packet[2:])
		fmt.Printf("Y: %v (shift: %v)\n", f, shift)
		//printBinary(p.packet[3:10])
		//fmt.Printf("s%v %v\n", shift, f)
		p.lastY = nv
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
	var c int
	nv := float64(binary.BigEndian.Uint32([]byte{0x00, out[0], out[1], out[2]}) >> shift)
	i := nv / 8.332
	if shift == 5 {
		c = 0
		i -= 3730
	} else if p.packet[3] >= 8 || p.packet[3] == 7 && p.packet[4] >= 128 {
		c = 1
	} else {
		c = 2
		i += 4135
	}
	_ = c
	if nv != p.lastX {
		//printBinary(p.packet[3:10])
		//fmt.Printf("sh%v c%v %v -> %v (%v)\n", shift, c, nv, math.Round(i*10)/10, p.lastX-nv)
		p.lastX = nv
	}
	return i
}

func (p *CarPosPacket) isLowY() bool {
	yHeader := binary.BigEndian.Uint16(p.packet[3:5])
	return int16(yHeader) <= 1941
}

func bitMask(n int) uint32 {
	out := 0x00
	for i := 0; i < n; i++ {
		out = out | (0x01 << uint(i))
	}
	return uint32(out)
}
