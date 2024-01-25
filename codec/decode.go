package codec

import (
	"encoding/binary"
	"fmt"
	"log/slog"
)

// convert raw bytes read from the file or socket to the unsigned ints sent by LIRC
func convertRawToLirc(rawData []byte) []uint32 {
	// round down to even 4 bytes
	n := len(rawData) - len(rawData)%4
	data := make([]uint32, 0, n)
	for i := 0; i < n; i = i + 4 {
		data = append(data, binary.LittleEndian.Uint32(rawData[i:i+4]))
	}
	return data
}

// round the value of pulses and spaces to the expected timings used by the Panasonic IR RC.
func roundToPanasonicIrTimings(v uint32) uint32 {
	mode2 := v & l_LIRC_MODE2_MASK
	length := v & l_LIRC_VALUE_MASK

	switch mode2 {
	case l_LIRC_MODE2_SPACE:
		// first make a nice try to find values close to the expected
		for _, t := range l_PANASONIC_IR_SPACE_TIMINGS() {
			if t-l_PANASONIC_TIMING_SPREAD < length && length < t+l_PANASONIC_TIMING_SPREAD {
				return t
			}
		}
		// if the previous didn't work, try a really coarse categorization
		// expected values are 435, 1300, 1750, or 10000
		// if length < 900 {
		// 	return l_PANASONIC_SPACE_0
		// } else if length < 1550 {
		// 	return l_PANASONIC_SPACE_1
		// } else if length < 2200 {
		// 	return l_PANASONIC_FRAME_MARK2_SPACE
		// } else if length > 9000 && length < l_PANASONIC_SPACE_OUTLIER {
		// 	return l_PANASONIC_SEPARATOR
		// }
	case l_LIRC_MODE2_PULSE:
		// first make a nice try to find values close to the expected
		for _, t := range l_PANASONIC_IR_PULSE_TIMINGS() {
			if t-l_PANASONIC_TIMING_SPREAD < length && length < t+l_PANASONIC_TIMING_SPREAD {
				return t
			}
		}
		// if the previous didn't work, try a really coarse categorization
		// expected values are 435 or 3500
		// if length < 900 {
		// 	return l_PANASONIC_PULSE
		// } else if length > 2500 && length < l_PANASONIC_PULSE_OUTLIER {
		// 	return l_PANASONIC_FRAME_MARK1_PULSE
		// }
	}
	return length
}

// Clean up the LIRC unsigned int data, by rounding pulses and spaces to the expected values,
// and filtering out all unexpected mode2 types.
func filterLircAsPanasonic(lircItem uint32) (bool, uint32) {
	switch lircItem & l_LIRC_MODE2_MASK {
	case l_LIRC_MODE2_SPACE:
		sp := roundToPanasonicIrTimings(lircItem)
		// discard long spaces that are not part of the protocol
		if sp >= l_PANASONIC_SPACE_OUTLIER {
			return false, 0
		}
		return true, sp | l_LIRC_MODE2_SPACE
	case l_LIRC_MODE2_PULSE:
		pu := roundToPanasonicIrTimings(lircItem)
		// discard long pulses that are not part of the protocol
		if pu >= l_PANASONIC_PULSE_OUTLIER {
			return false, 0
		}
		return true, pu | l_LIRC_MODE2_PULSE
	case l_LIRC_MODE2_TIMEOUT:
		// this basically means that we've reached the end of a transmission
		return true, lircItem
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
	return fmt.Sprintf("pos %d: %s (status %d)", state.pos, state.description, state.status)
}

func findStartOfPanasonicFrame(data []uint32) (int, error) {
	// find start of frame
	for i := 0; i < len(data)-1; i++ {
		if data[i] == (l_LIRC_MODE2_PULSE|l_PANASONIC_FRAME_MARK1_PULSE) && data[i+1] == (l_LIRC_MODE2_SPACE|l_PANASONIC_FRAME_MARK2_SPACE) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no start of frame found")
}

func isTimeout(v uint32) bool {
	return v&l_LIRC_MODE2_MASK == l_LIRC_MODE2_TIMEOUT
}

func findEndOfData(data []uint32, pos int) (eod int, timeoutFound bool) {
	// find timeout, if any
	for i := pos; i < len(data); i++ {
		if isTimeout(data[i]) {
			return i, true
		}
	}
	return len(data), false
}

func skipToken(lircData []uint32, pos int, mode2, value uint32) parseState {
	if pos >= len(lircData) {
		return parseState{pos, PARSE_END_OF_DATA, fmt.Sprintf("reached end-of-data while parsing, pos=%d", pos)}
	}
	d := lircData[pos]
	if d&l_LIRC_MODE2_MASK != mode2 {
		return parseState{pos, PARSE_UNEXPECTED_MODE2, fmt.Sprintf("expected mode2 %#08x, found %#08x", mode2, d&l_LIRC_MODE2_MASK)}
	}
	if d&l_LIRC_VALUE_MASK != value {
		return parseState{pos, PARSE_UNEXPECTED_VALUE, fmt.Sprintf("expected value %d, found %d", value, d&l_LIRC_VALUE_MASK)}
	}
	return parseState{pos + 1, PARSE_OK, "skipped expected token"}
}

func skipSpace(lircData []uint32, pos int, expectedSpace uint32) parseState {
	return skipToken(lircData, pos, l_LIRC_MODE2_SPACE, expectedSpace)
}

func skipPulse(lircData []uint32, pos int, expectedPulse uint32) parseState {
	return skipToken(lircData, pos, l_LIRC_MODE2_PULSE, expectedPulse)
}

func readSpace(lircData []uint32, pos int) (uint32, parseState) {
	if pos >= len(lircData) {
		return 0, parseState{pos, PARSE_END_OF_DATA, fmt.Sprintf("reached end-of-data while parsing, pos=%d", pos)}
	}
	d := lircData[pos]
	if d&l_LIRC_MODE2_MASK != l_LIRC_MODE2_SPACE {
		return 0, parseState{pos, PARSE_UNEXPECTED_MODE2, fmt.Sprintf("expected mode2 %#08x, found %#08x", l_LIRC_MODE2_SPACE, d&l_LIRC_MODE2_MASK)}
	}
	return d & l_LIRC_VALUE_MASK, parseState{pos + 1, PARSE_OK, "read a space"}
}

func appendPanasonicBit(space uint32, frame *Frame) error {
	var bit uint
	switch space {
	case l_PANASONIC_SPACE_0:
		bit = 0
	case l_PANASONIC_SPACE_1:
		bit = 1
	default:
		return fmt.Errorf("cannot translate space length to bit: %d", space)
	}
	(*frame).AppendBit(bit)
	return nil
}

func parsePanasonicFrame(lircData []uint32, pos int, nBits int, frame *Frame, options *receiverOptions) parseState {
	state := skipPulse(lircData, pos, l_PANASONIC_FRAME_MARK1_PULSE)
	if state.status != PARSE_OK {
		slog.Debug("mark1 pulse not found")
		return state
	}
	state = skipSpace(lircData, state.pos, l_PANASONIC_FRAME_MARK2_SPACE)
	if state.status != PARSE_OK {
		slog.Debug("mark2 space not found")
		return state
	}
	for i := 0; i < nBits; i++ {
		var space uint32
		state = skipPulse(lircData, state.pos, l_PANASONIC_PULSE)
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
	state = skipPulse(lircData, state.pos, l_PANASONIC_PULSE)
	if state.status != PARSE_OK {
		return state
	}
	return state
}

func readPanasonicMessage(lircData []uint32, options *receiverOptions) (*Message, []uint32, parseState) {
	// slog.Debug("parse data", "items", len(lircData), "required", l_PANASONIC_LIRC_ITEMS)
	start, err := findStartOfPanasonicFrame(lircData)
	if err != nil {
		return nil, lircData, parseState{0, PARSE_MISSING_START_OF_FRAME, "start of frame not found"}
	}
	end, foundTimeout := findEndOfData(lircData, start)
	// slog.Debug("findEndOfData", "start", start, "end", end, "timeout", foundTimeout)
	if foundTimeout && end-start < l_PANASONIC_LIRC_ITEMS {
		// we found an end-of-trasmission but it can't be a full message
		slog.Debug("discarding truncated message")
		return nil, lircData[end:], parseState{end, PARSE_NOT_ENOUGH_DATA, "truncated message"}
	}
	if end-start < l_PANASONIC_LIRC_ITEMS {
		// read more until the minimum required bytes in a message have been received
		return nil, lircData[start:], parseState{start, PARSE_NOT_ENOUGH_DATA, "expecting more data"}
	}

	msg := NewMessage()

	state := parsePanasonicFrame(lircData[:end], start, l_PANASONIC_BITS_FRAME1, &msg.Frame1, options)
	if state.status != PARSE_OK {
		return nil, lircData[state.pos+1:], state
	}
	state = skipSpace(lircData[:end], state.pos, l_PANASONIC_SEPARATOR)
	if state.status != PARSE_OK {
		return nil, lircData[state.pos+1:], state
	}
	state = parsePanasonicFrame(lircData[:end], state.pos, l_PANASONIC_BITS_FRAME2, &msg.Frame2, options)
	if state.status != PARSE_OK {
		return nil, lircData[state.pos+1:], state
	}
	return msg, lircData[state.pos:], parseState{state.pos, PARSE_OK, "parsed a complete message"}
}
