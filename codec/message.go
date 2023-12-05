package codec

import (
	"fmt"
	"math/big"
	"strings"
)

type Frame interface {
	AppendBit(bit uint) (nBits int)
	GetValue(bitIndex uint, numberOfBits uint) (value uint)
	SetValue(value uint, bitIndex uint, numberOfBits uint) Frame
	GetChecksum() byte
	ComputeChecksum() byte
	VerifyChecksum() bool
	SetChecksum()
	ToTraceString() (traceS, posS string)
	ToBitStream() string
	ToByteString() string
	Equal(other Frame) bool
	ToLirc(b *LircBuffer)
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
	return &Message{
		Frame1: &BitSet{bs1.SetBytes(p_PANASONIC_FRAME1()), l_PANASONIC_BITS_FRAME1},
		Frame2: &BitSet{bs2.SetBytes(p_PANASONIC_FRAME2()), l_PANASONIC_BITS_FRAME2},
	}
}

func (msg *Message) ToLirc() *LircBuffer {
	b := NewLircBuffer()
	msg.Frame1.ToLirc(b)
	b.FrameSpace()
	msg.Frame2.ToLirc(b)
	return b
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
func (f *BitSet) GetValue(bitIndex uint, numberOfBits uint) (value uint) {
	// copy bits from bitset to value
	for i := 0; i < int(numberOfBits); i++ {
		value = value | (f.bits.Bit(i+int(bitIndex)) << i)
	}
	return value
}

// Sets a value at a certain position in the bit stream. Returns the BitSet.
func (f *BitSet) SetValue(value uint, bitIndex uint, numberOfBits uint) Frame {
	var bit uint
	// copy bits from value to bitset, overwriting any bits already there
	for i := 0; i < int(numberOfBits); i++ {
		bit, value = uint(value&1), value>>1
		f.bits.SetBit(f.bits, i+int(bitIndex), bit)
	}
	return f
}

func (f *BitSet) ToTraceString() (traceS, posS string) {
	posS = ""
	for i := 0; i < f.n; i += 8 {
		posS = fmt.Sprintf("%9d", i) + posS
	}
	posS = "       " + posS
	return fmt.Sprintf("%4d/%d %s %t", f.n, f.n%8, f.ToBitStream(), f.VerifyChecksum()), posS
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

	// if we have the full number of bytes, don't include byte 0 (the checksum) in the computation
	start := 0
	expectedBytes := (f.n + 7) / 8
	if ln >= expectedBytes {
		start = 1
	}

	var sum byte = 0
	for i := start; i < ln; i++ {
		sum += b[i]
	}
	return sum
}

func (f *BitSet) VerifyChecksum() bool {
	return f.ComputeChecksum() == f.GetChecksum()
}

func (f *BitSet) SetChecksum() {
	cs := f.ComputeChecksum()
	f.SetValue(uint(cs), uint(f.n-p_PANASONIC_CHECKSUM_BITS), p_PANASONIC_CHECKSUM_BITS)
}

func (f *BitSet) Equal(other Frame) bool {
	o := other.(*BitSet)
	return f.n == o.n && f.bits.Cmp(o.bits) == 0
}

func (f *BitSet) ToLirc(b *LircBuffer) {
	b.BeginFrame()
	for i := 0; i < f.n; i++ {
		b.AddBit(f.bits.Bit(i))
	}
	b.EndFrame()
}
