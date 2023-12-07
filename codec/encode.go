package codec

import "encoding/binary"

type LircBuffer struct {
	buf []uint32
}

func NewLircBuffer() *LircBuffer {
	return &LircBuffer{make([]uint32, 0, l_PANASONIC_LIRC_ITEMS)}
}

func (b *LircBuffer) BeginFrame() {
	b.addPulse(l_PANASONIC_FRAME_MARK1_PULSE)
	b.addSpace(l_PANASONIC_FRAME_MARK2_SPACE)
}

func (b *LircBuffer) EndFrame() {
	b.addPulse(l_PANASONIC_PULSE)
}

func (b *LircBuffer) FrameSpace() {
	b.addSpace(l_PANASONIC_SEPARATOR)
}

func (b *LircBuffer) AddBit(bit uint) {
	b.addPulse(l_PANASONIC_PULSE)
	if bit == 0 {
		b.addSpace(l_PANASONIC_SPACE_0)
	} else {
		b.addSpace(l_PANASONIC_SPACE_1)
	}
}

func (b *LircBuffer) addPulse(length uint32) {
	b.buf = append(b.buf, length|l_LIRC_MODE2_PULSE)
}

func (b *LircBuffer) addSpace(length uint32) {
	b.buf = append(b.buf, length|l_LIRC_MODE2_SPACE)
}

func (b *LircBuffer) ToBytes() (bytes []byte) {
	bytes = make([]byte, 0, len(b.buf)*4)
	for i := 0; i < len(b.buf); i++ {
		bytes = binary.LittleEndian.AppendUint32(bytes, b.buf[i])
	}
	return bytes
}
