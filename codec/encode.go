package codec

import (
	"encoding/binary"
	"fmt"
	"rpi_panasonic_inverter_rc/common"
)

type LircBuffer struct {
	buf []uint32
}

func NewLircBuffer() *LircBuffer {
	return &LircBuffer{make([]uint32, 0, common.L_PANASONIC_LIRC_ITEMS)}
}

func (b *LircBuffer) BeginFrame() {
	b.addPulse(common.L_PANASONIC_FRAME_MARK1_PULSE)
	b.addSpace(common.L_PANASONIC_FRAME_MARK2_SPACE)
}

func (b *LircBuffer) EndFrame() {
	b.addPulse(common.L_PANASONIC_PULSE)
}

func (b *LircBuffer) FrameSpace() {
	b.addSpace(common.L_PANASONIC_SEPARATOR)
}

func (b *LircBuffer) AddBit(bit uint) {
	b.addPulse(common.L_PANASONIC_PULSE)
	if bit == 0 {
		b.addSpace(common.L_PANASONIC_SPACE_0)
	} else {
		b.addSpace(common.L_PANASONIC_SPACE_1)
	}
}

func (b *LircBuffer) addPulse(length uint32) {
	b.buf = append(b.buf, length|common.L_LIRC_MODE2_PULSE)
}

func (b *LircBuffer) addSpace(length uint32) {
	b.buf = append(b.buf, length|common.L_LIRC_MODE2_SPACE)
}

func (b *LircBuffer) ToBytes() (bytes []byte) {
	bytes = make([]byte, 0, len(b.buf)*4)
	for i := 0; i < len(b.buf); i++ {
		bytes = binary.LittleEndian.AppendUint32(bytes, b.buf[i])
	}
	return bytes
}

func (b *LircBuffer) PrintLircBuffer() {
	for _, code := range b.buf {
		printLircData("LircBuffer", code)
	}
}

func printLircData(label string, d uint32) {
	v := d & common.L_LIRC_VALUE_MASK
	fmt.Printf("%s\t", label)
	switch d & common.L_LIRC_MODE2_MASK {
	case common.L_LIRC_MODE2_SPACE:
		fmt.Printf("space\t%d\n", v)
	case common.L_LIRC_MODE2_PULSE:
		fmt.Printf("pulse\t%d\n", v)
	case common.L_LIRC_MODE2_FREQUENCY:
		fmt.Printf("frequencyt%d\n", v)
	case common.L_LIRC_MODE2_TIMEOUT:
		fmt.Printf("timeout\t%d\n", v)
	case common.L_LIRC_MODE2_OVERFLOW:
		fmt.Printf("overflow\t%d\n", v)
	}
}
