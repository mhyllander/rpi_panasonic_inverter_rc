package codec

import (
	"fmt"
	"math/big"
	"strings"
)

type Frame interface {
	AppendBit(bit uint) (nBits int)
	GetValue(bitIndex uint, numberOfBits uint) uint64
	SetValue(value uint64, bitIndex uint, numberOfBits uint) Frame
	GetChecksum() byte
	ComputeChecksum() byte
	VerifyChecksum() bool
	ToTraceString() (traceS, posS string)
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

// return an empty message suitable for receiving
func NewMessage() *Message {
	msg := Message{
		Frame1: &BitSet{big.NewInt(0), 0},
		Frame2: &BitSet{big.NewInt(0), 0},
	}
	return &msg
}

// return an initialized message suitable for sending
func InitializedMessage() *Message {
	var bs1, bs2 big.Int
	bs1.SetBytes(PANASONIC_FRAME1())
	bs2.SetBytes(PANASONIC_FRAME2())
	msg := Message{
		Frame1: &BitSet{&bs1, PANASONIC_BITS_FRAME1},
		Frame2: &BitSet{&bs2, PANASONIC_BITS_FRAME2},
	}
	return &msg
}

// Appends a bit to the bitstream, and returns the number of bits
// The bits are sent in little endian order. By appending each bit to the left in the Int,
// we maintain the abstraction that the first bit sent is at index 0, while at the same time
// converting to big endian.
func (f *BitSet) AppendBit(bit uint) (nBits int) {
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

// Sets a value at a certain position in the bit stream. Returns the BitSet.
func (f *BitSet) SetValue(value uint64, bitIndex uint, numberOfBits uint) Frame {
	var bBits big.Int
	bBits.SetUint64(value)

	// copy bits from value to bitset, overwriting any bits already there
	for i := int(0); i < int(numberOfBits); i++ {
		f.bits.SetBit(f.bits, i, bBits.Bit(i))
	}

	return f
}

func (f *BitSet) ToTraceString() (traceS, posS string) {
	checksumOK := f.VerifyChecksum()
	posS = ""
	for i := 0; i < f.n; i += 8 {
		posS = fmt.Sprintf("%9d", i) + posS
	}
	posS = "       " + posS
	return fmt.Sprintf("%4d/%d %s %t", f.n, f.n%8, f.ToBitStream(), checksumOK), posS
}

func (f *BitSet) ToBitStream() string {
	if f.n == 0 {
		return ""
	}
	b := make([]byte, (f.n+7)/8)
	f.bits.FillBytes(b)
	return fmt.Sprintf("%08b", b)
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

func (f *BitSet) GetChecksum() byte {
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
	return f.ComputeChecksum() == f.GetChecksum()
}
