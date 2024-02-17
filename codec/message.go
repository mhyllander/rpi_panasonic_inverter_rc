package codec

import (
	"fmt"
	"math/big"
	"rpi_panasonic_inverter_rc/common"
)

type Frame interface {
	AppendBit(bit uint) (nBits int)
	GetValue(bitIndex uint, numberOfBits uint) (value uint)
	SetValue(value uint, bitIndex uint, numberOfBits uint) Frame
	GetChecksum() byte
	ComputeChecksum() byte
	VerifyChecksum() bool
	SetChecksum()
	ToVerboseString() (verboseS, posS string)
	ToBitStream() string
	ToByteString() string
	Equal(other Frame) bool
	ToLirc(b *LircBuffer)
}

type Message struct {
	Frame1 Frame
	Frame2 Frame
}

// Create an empty Message, using BitSet as the Frame representation. Suitable for receiving a message.
func NewMessage() *Message {
	msg := Message{
		Frame1: &BitSet{big.NewInt(0), 0},
		Frame2: &BitSet{big.NewInt(0), 0},
	}
	return &msg
}

// Create an initialized Message, using BitSet as the Frame representation. Suitable for sending a message.
func InitializedMessage() *Message {
	var bs1, bs2 big.Int
	return &Message{
		Frame1: &BitSet{bs1.SetBytes(common.P_PANASONIC_FRAME1()), common.L_PANASONIC_BITS_FRAME1},
		Frame2: &BitSet{bs2.SetBytes(common.P_PANASONIC_FRAME2()), common.L_PANASONIC_BITS_FRAME2},
	}
}

func (msg *Message) PrintMessage() {
	t1, p1 := msg.Frame1.ToVerboseString()
	t2, p2 := msg.Frame2.ToVerboseString()

	fmt.Printf("Message as bit stream (first and least significant bit to the right)\n")
	fmt.Printf("   %s\n%d: %s\n", p1, 1, t1)
	fmt.Printf("   %s\n%d: %s\n", p2, 2, t2)
}

func (msg *Message) PrintByteRepresentation() {
	fmt.Println("Byte representation:")
	fmt.Printf("  %d: %s\n", 1, msg.Frame1.ToByteString())
	fmt.Printf("  %d: %s\n", 2, msg.Frame2.ToByteString())
}

func (msg *Message) ToLirc() *LircBuffer {
	b := NewLircBuffer()
	msg.Frame1.ToLirc(b)
	b.FrameSpace()
	msg.Frame2.ToLirc(b)
	return b
}
