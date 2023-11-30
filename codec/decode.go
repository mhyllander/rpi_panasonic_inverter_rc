package codec

import (
	"encoding/binary"
	"fmt"
)

// convert raw bytes read from the file or socket to the unsigned ints sent by LIRC
func convertRawToLirc(rawData []byte) []uint32 {
	data := make([]uint32, 0, len(rawData)/4+1)
	for i := 0; i < len(rawData); i = i + 4 {
		uintSlice := rawData[i:(i + 4)]
		d := binary.LittleEndian.Uint32(uintSlice)
		data = append(data, d)
	}
	return data
}

// round the value of pulses and spaces to the expected timings used by the Panasonic IR RC.
func roundToPanasonicIrTimings(v uint32) uint32 {
	timings := PANASONIC_IR_TIMINGS()
	for _, t := range timings {
		if t-LIRC_TIMING_SPREAD < v && v < t+LIRC_TIMING_SPREAD {
			return t
		}
	}
	return v
}

// Clean up the LIRC unsigned int data, by rounding pulses and spaces to the expected values,
// and filtering out all unexpected mode2 types.
func filterLircAsPanasonic(lircItem uint32) (bool, uint32) {
	v := lircItem & LIRC_VALUE_MASK
	switch lircItem & LIRC_MODE2_MASK {
	case LIRC_MODE2_SPACE:
		sp := roundToPanasonicIrTimings(v)
		// discard long spaces that are not part of the protocol
		if sp > PANASONIC_SPACE_OUTLIER {
			return false, 0
		}
		return true, sp | LIRC_MODE2_SPACE
	case LIRC_MODE2_PULSE:
		return true, roundToPanasonicIrTimings(v) | LIRC_MODE2_PULSE
	default:
		// discard other data
	}
	return false, 0
}

const (
	PARSE_OK = iota + 1
	PARSE_MISSING_START_OF_FRAME
	PARSE_UNEXPECTED_MODE2
	PARSE_UNEXPECTED_VALUE
	PARSE_NOT_ENOUGH_DATA
	PARSE_END_OF_DATA
	PARSE_ERROR
)

type parseState struct {
	pos         int
	status      int
	description string
}

func (state parseState) Error() string {
	return fmt.Sprintf("%s (status %d)", state.description, state.status)
}

func findStartOfPanasonicFrame(data []uint32) (int, error) {
	// find start of frame
	for i := 0; i < len(data)-1; i++ {
		if data[i] == (LIRC_MODE2_PULSE|PANASONIC_FRAME_MARK1) && data[i+1] == (LIRC_MODE2_SPACE|PANASONIC_FRAME_MARK2) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no start of frame found")
}

func skipToken(lircData []uint32, pos int, mode2, value uint32) parseState {
	if pos >= len(lircData) {
		return parseState{pos, PARSE_END_OF_DATA, fmt.Sprintf("reached end-of-data while parsing, pos=%d", pos)}
	}
	d := lircData[pos]
	if d&LIRC_MODE2_MASK != mode2 {
		return parseState{pos, PARSE_UNEXPECTED_MODE2, fmt.Sprintf("expected mode2 %#08x, found %#08x", mode2, d&LIRC_MODE2_MASK)}
	}
	if d&LIRC_VALUE_MASK != value {
		return parseState{pos, PARSE_UNEXPECTED_VALUE, fmt.Sprintf("expected value %d, found %d", value, d&LIRC_VALUE_MASK)}
	}
	return parseState{pos + 1, PARSE_OK, "skipped expected token"}
}

func skipSpace(lircData []uint32, pos int, expectedSpace uint32) parseState {
	return skipToken(lircData, pos, LIRC_MODE2_SPACE, expectedSpace)
}

func skipPulse(lircData []uint32, pos int, expectedPulse uint32) parseState {
	return skipToken(lircData, pos, LIRC_MODE2_PULSE, expectedPulse)
}

func readSpace(lircData []uint32, pos int) (uint32, parseState) {
	if pos >= len(lircData) {
		return 0, parseState{pos, PARSE_END_OF_DATA, fmt.Sprintf("reached end-of-data while parsing, pos=%d", pos)}
	}
	d := lircData[pos]
	if d&LIRC_MODE2_MASK != LIRC_MODE2_SPACE {
		return 0, parseState{pos, PARSE_UNEXPECTED_MODE2, fmt.Sprintf("expected mode2 %#08x, found %#08x", LIRC_MODE2_SPACE, d&LIRC_MODE2_MASK)}
	}
	return d & LIRC_VALUE_MASK, parseState{pos + 1, PARSE_OK, "read a space"}
}

func appendPanasonicBit(space uint32, frame *Frame) error {
	var bit uint
	switch space {
	case PANASONIC_SPACE_0:
		bit = 0
	case PANASONIC_SPACE_1:
		bit = 1
	default:
		return fmt.Errorf("cannot translate space length to bit: %d", space)
	}
	(*frame).AppendBit(bit)
	return nil
}

func parsePanasonicFrame(lircData []uint32, pos int, nBits int, frame *Frame, options *ReaderOptions) parseState {
	state := skipPulse(lircData, pos, PANASONIC_FRAME_MARK1)
	if state.status != PARSE_OK {
		if options.Trace {
			fmt.Println("mark1 pulse not found")
		}
		return state
	}
	state = skipSpace(lircData, state.pos, PANASONIC_FRAME_MARK2)
	if state.status != PARSE_OK {
		if options.Trace {
			fmt.Println("mark2 space not found")
		}
		return state
	}
	for i := 0; i < nBits; i++ {
		var space uint32
		state = skipPulse(lircData, state.pos, PANASONIC_PULSE)
		if state.status != PARSE_OK {
			return state
		}
		space, state = readSpace(lircData, state.pos)
		if state.status != PARSE_OK {
			return state
		}
		err := appendPanasonicBit(space, frame)
		if err != nil {
			return parseState{pos, PARSE_ERROR, err.Error()}
		}
	}
	state = skipPulse(lircData, state.pos, PANASONIC_PULSE)
	if state.status != PARSE_OK {
		return state
	}
	return state
}

func readPanasonicMessage(lircData []uint32, options *ReaderOptions) (*Message, []uint32, parseState) {
	if len(lircData) < PANASONIC_LIRC_ITEMS {
		// read more until the minimum required bytes in a message have been received
		return nil, lircData, parseState{0, PARSE_NOT_ENOUGH_DATA, "expecting more data"}
	}
	start, err := findStartOfPanasonicFrame(lircData)
	if err != nil {
		return nil, lircData, parseState{0, PARSE_MISSING_START_OF_FRAME, "start of frame not found"}
	}

	msg := NewMessage()

	state := parsePanasonicFrame(lircData, start, PANASONIC_BITS_FRAME1, &msg.Frame1, options)
	if state.status != PARSE_OK {
		return nil, lircData, state
	}
	state = skipSpace(lircData, state.pos, PANASONIC_SEPARATOR)
	if state.status != PARSE_OK {
		return nil, lircData, state
	}
	state = parsePanasonicFrame(lircData, state.pos, PANASONIC_BITS_FRAME2, &msg.Frame2, options)
	if state.status != PARSE_OK {
		return nil, lircData, state
	}
	return msg, lircData[state.pos:], parseState{state.pos, PARSE_OK, "parsed a complete message"}
}
