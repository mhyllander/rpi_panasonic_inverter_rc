package codec

import (
	"fmt"
	"math/big"
	"math/bits"
	"strings"
)

type Frame interface {
	AppendBit(bit uint) int
	GetValue(bitIndex uint, numberOfBits uint) uint64
	GetCheckSum() byte
	ComputeChecksum() byte
	VerifyChecksum() bool
	ToTraceString() string
	ToBitStream() string
	ToByteString() string
}

type Message struct {
	Frame1 Frame
	Frame2 Frame
}

type BitSet struct {
	bits *big.Int
	n    int
}

func NewMessage() *Message {
	msg := Message{
		Frame1: &BitSet{big.NewInt(0), 0},
		Frame2: &BitSet{big.NewInt(0), 0},
	}
	return &msg
}

// Appends a bit to the bitstream, and returns the number of bits
// The bits are sent in little endian order. By appending each bit to the left in the Int,
// we maintain the abstraction that the first bit sent is at index 0, while at the same time
// converting to big endian.
func (f *BitSet) AppendBit(bit uint) int {
	f.bits.SetBit(f.bits, f.n, bit)
	f.n++
	return f.n
}

// Retrieves a value from a certain position in the bit stream.
// Since the bits were received with least significant bit first, this means that
// the least significant bit is rightmost, which is convenient when we want to get
// a value at a certain index. We simply shift the bits right so that the first bit
// is at index 0, then apply the mask.
func (f *BitSet) GetValue(bitIndex uint, numberOfBits uint) uint64 {
	var bMask, bBits big.Int
	mask := (1 << numberOfBits) - 1
	bMask.SetUint64(uint64(mask))
	return bBits.Rsh(f.bits, bitIndex).And(&bBits, &bMask).Uint64()
}

func (f BitSet) ToTraceString() string {
	checksumOK := f.VerifyChecksum()
	return fmt.Sprintf("%4d/%d %s %t", f.n, f.n%8, f.ToBitStream(), checksumOK)
}

func (f *BitSet) ToBitStream() string {
	if f.n == 0 {
		return ""
	}
	b := make([]byte, (f.n+7)/8)
	f.bits.FillBytes(b)
	s := ""
	for i := len(b) - 1; i >= 0; i-- {
		s += fmt.Sprintf("%08b ", bits.Reverse8(b[i]))
	}
	s, _ = strings.CutSuffix(s, " ")
	return "[" + s + "]"
}

func (f *BitSet) ToByteString() string {
	if f.n == 0 {
		return ""
	}
	b := make([]byte, (f.n+7)/8)
	f.bits.FillBytes(b)
	s := ""
	for i := 0; i < len(b); i++ {
		s += fmt.Sprintf("%#08b, ", b[i])
	}
	s, _ = strings.CutSuffix(s, ", ")
	return "{" + s + "}"
}

func (f *BitSet) GetCheckSum() byte {
	if f.n == 0 {
		return 0
	}
	return f.bits.Bytes()[0]
}

func (f *BitSet) ComputeChecksum() byte {
	if f.n == 0 {
		return 0
	}
	b := f.bits.Bytes()
	ln := len(b)
	var sum byte = 0
	// do not include byte 0 (the received checksum)
	for i := 1; i < ln; i++ {
		sum += b[i]
	}
	return sum
}

func (f *BitSet) VerifyChecksum() bool {
	return f.ComputeChecksum() == f.GetCheckSum()
}
