package codec

import (
	"encoding/binary"
	"fmt"

	"rpi_panasonic_inverter_rc/codecbase"
)

type LircBuffer struct {
	buf []uint32
}

func NewLircBuffer() *LircBuffer {
	return &LircBuffer{make([]uint32, 0, codecbase.L_PANASONIC_LIRC_ITEMS)}
}

func (b *LircBuffer) BeginFrame() {
	b.addPulse(codecbase.L_PANASONIC_FRAME_MARK1_PULSE)
	b.addSpace(codecbase.L_PANASONIC_FRAME_MARK2_SPACE)
}

func (b *LircBuffer) EndFrame() {
	b.addPulse(codecbase.L_PANASONIC_PULSE)
}

func (b *LircBuffer) FrameSpace() {
	b.addSpace(codecbase.L_PANASONIC_SEPARATOR)
}

func (b *LircBuffer) AddBit(bit uint) {
	b.addPulse(codecbase.L_PANASONIC_PULSE)
	if bit == 0 {
		b.addSpace(codecbase.L_PANASONIC_SPACE_0)
	} else {
		b.addSpace(codecbase.L_PANASONIC_SPACE_1)
	}
}

func (b *LircBuffer) addPulse(length uint32) {
	b.buf = append(b.buf, length|codecbase.L_LIRC_MODE2_PULSE)
}

func (b *LircBuffer) addSpace(length uint32) {
	b.buf = append(b.buf, length|codecbase.L_LIRC_MODE2_SPACE)
}

func (b *LircBuffer) ToBytes() (bytes []byte) {
	bytes = make([]byte, 0, len(b.buf)*4)
	for i := 0; i < len(b.buf); i++ {
		bytes = binary.LittleEndian.AppendUint32(bytes, b.buf[i])
	}
	return bytes
}

func (b *LircBuffer) ToMode2Lirc() []string {
	s := make([]string, 0, 500)
	for _, v := range b.buf {
		if v&codecbase.L_LIRC_MODE2_MASK == codecbase.L_LIRC_MODE2_PULSE {
			s = append(s, fmt.Sprintf("+%d", v&codecbase.L_LIRC_VALUE_MASK))
		} else if v&codecbase.L_LIRC_MODE2_MASK == codecbase.L_LIRC_MODE2_SPACE {
			s = append(s, fmt.Sprintf("-%d", v&codecbase.L_LIRC_VALUE_MASK))
		}
	}
	return s
}

func (b *LircBuffer) PrintLircBuffer() {
	for _, code := range b.buf {
		printLircData("LircBuffer", code)
	}
}

func printLircData(label string, d uint32) {
	v := d & codecbase.L_LIRC_VALUE_MASK
	fmt.Printf("%s\t", label)
	switch d & codecbase.L_LIRC_MODE2_MASK {
	case codecbase.L_LIRC_MODE2_SPACE:
		fmt.Printf("space\t%d\n", v)
	case codecbase.L_LIRC_MODE2_PULSE:
		fmt.Printf("pulse\t%d\n", v)
	case codecbase.L_LIRC_MODE2_FREQUENCY:
		fmt.Printf("frequencyt%d\n", v)
	case codecbase.L_LIRC_MODE2_TIMEOUT:
		fmt.Printf("timeout\t%d\n", v)
	case codecbase.L_LIRC_MODE2_OVERFLOW:
		fmt.Printf("overflow\t%d\n", v)
	}
}
