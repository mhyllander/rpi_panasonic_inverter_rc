package codec

import (
	"fmt"
	"math/bits"
)

type Frame struct {
	Data  []byte
	NBits int
}

func (f Frame) ToTraceString() string {
	checksumOK, checkSum := f.VerifyChecksum()
	return fmt.Sprintf("%4d/%d %08b %t(%08b)", f.NBits, f.NBits%8, f.Data, checksumOK, checkSum)
}

func (f Frame) ToBinaryString() string {
	return fmt.Sprintf("%08b)", f.Data)
}

func (f Frame) ComputeChecksum() byte {
	var sum byte = 0
	ln := len(f.Data) - 1
	for i := 0; i < ln; i++ {
		sum += f.Data[i]
	}
	return sum
}

func (f Frame) VerifyChecksum() (bool, byte) {
	sum := f.ComputeChecksum()
	return sum == f.Data[len(f.Data)-1], sum
}

// Assemble a value using bits from one byte.
func (f Frame) ValueFromOneByte(b1 int, mask uint8) uint8 {
	shift := bits.TrailingZeros8(mask)
	return f.Data[b1] & mask >> shift
}

// Assemble a value using bits from two bytes.
// The bits from byte 1 are least significant, the bits from byte 2 are most significant.
func (f Frame) ValueFromTwoBytes(b1 int, mask1 uint8, b2 int, mask2 uint8) uint16 {
	bits1 := bits.OnesCount8(mask1)
	shift1 := bits.TrailingZeros8(mask1)
	shift2 := bits.TrailingZeros8(mask2)

	v1 := uint16(f.Data[b1] & mask1 >> shift1)
	v2 := uint16(f.Data[b2] & mask2 >> shift2)
	return v2<<bits1 | v1
}
