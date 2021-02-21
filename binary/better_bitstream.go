package binary

import (
	"fmt"
	"github.com/WorldUnitedNFS/freeroam/math"
)

// Bitstream allows to read individual bits from an array of bytes.
type Bitstream struct {
	data       []byte
	dataOffset int
	bitsInto   uint
	bufferUsed uint
	buffer     uint32
}

// NewBitstream creates a Bitstream from an array of bytes.
func NewBitstream(data []byte) *Bitstream {
	bs := &Bitstream{
		data:       data,
		bitsInto:   0,
		bufferUsed: 0,
		dataOffset: 0,
	}

	return bs
}

func (bs *Bitstream) IsNextBitSet() bool {
	b, e := bs.ReadBit()

	if e != nil {
		return false
	}

	return b == true
}

func (bs *Bitstream) GetData() []byte {
	return bs.data
}

func (bs *Bitstream) GetDataAtCurrentOffset() []byte {
	return bs.data[bs.dataOffset:]
}

func (bs *Bitstream) GetDataAtEnd() []byte {
	return bs.data[len(bs.data)-1:]
}

func (bs *Bitstream) GetSizeInBytes() int {
	return len(bs.data)
}

func (bs *Bitstream) ReadBit() (bool, error) {
	val, err := bs.ReadBits(1)

	if err != nil {
		return false, err
	}

	return val == 1, nil
}

func (bs *Bitstream) ReadBits(count uint) (uint32, error) {
	if count > 32 {
		return 0, fmt.Errorf("count must be less than or equal to 32")
	}

	if bs.bufferUsed < 0 || bs.bufferUsed > 32 {
		return 0, fmt.Errorf("internal state is corrupted: bufferUsed must be within range [0, 32]")
	}

	count = math.MinUI(count, bs.GetCountBitsLeft())

	if count == 0 {
		return 0, nil
	}

	if bs.bufferUsed < count {
		bs.fillBuffer()
	}

	bits := (bs.buffer >> (32 - count)) & (0xFFFFFFFF >> (32 - count))
	bs.updateBuffer(count)

	count += bs.bitsInto
	bs.dataOffset += int(count / 8)
	bs.bitsInto = count % 8

	if bs.dataOffset > len(bs.data) {
		return 0, fmt.Errorf("ran over data")
	}

	return bits, nil
}

func (bs *Bitstream) GetCountBitsLeft() uint {
	return uint(len(bs.data)-bs.dataOffset)*8 - bs.bitsInto
}

func (bs *Bitstream) Tell() uint {
	return uint(bs.dataOffset)*8 + bs.bitsInto
}

func (bs *Bitstream) fillBuffer() {
	bs.bufferUsed = math.MinUI(32, bs.GetCountBitsLeft())
	remainingBytes := len(bs.data) - bs.dataOffset

	if remainingBytes > 4 {
		bs.buffer = (uint32(bs.data[bs.dataOffset]) << (24 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+1]) << (16 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+2]) << (8 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+3]) << bs.bitsInto) |
			(uint32(bs.data[bs.dataOffset+4]) >> (8 - bs.bitsInto))
	} else if remainingBytes > 3 {
		bs.buffer = (uint32(bs.data[bs.dataOffset]) << (24 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+1]) << (16 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+2]) << (8 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+3]) << bs.bitsInto)
	} else if remainingBytes > 2 {
		bs.buffer = (uint32(bs.data[bs.dataOffset]) << (24 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+1]) << (16 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+2]) << (8 + bs.bitsInto))
	} else if remainingBytes > 1 {
		bs.buffer = (uint32(bs.data[bs.dataOffset]) << (24 + bs.bitsInto)) |
			(uint32(bs.data[bs.dataOffset+1]) << (16 + bs.bitsInto))
	} else if remainingBytes > 0 {
		bs.buffer = uint32(bs.data[bs.dataOffset]) << (24 + bs.bitsInto)
	}
}

func (bs *Bitstream) updateBuffer(count uint) {
	count = math.MinUI(count, bs.bufferUsed)
	bs.buffer <<= count
	bs.bufferUsed -= count
}
