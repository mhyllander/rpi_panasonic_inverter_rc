package codec

import (
	"encoding/binary"
	"fmt"
)

func convertRawToLirc(rawData []byte) []uint32 {
	data := make([]uint32, 0, len(rawData)/4+1)
	for i := 0; i < len(rawData); i = i + 4 {
		uintSlice := rawData[i:(i + 4)]
		d := binary.LittleEndian.Uint32(uintSlice)
		data = append(data, d)
	}
	return data
}

func roundToPanasonicIrTimings(v uint32) uint32 {
	timings := PANASONIC_IR_TIMINGS()
	for _, t := range timings {
		if t-LIRC_TIMING_SPREAD < v && v < t+LIRC_TIMING_SPREAD {
			return t
		}
	}
	return v
}

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

func findStartOfPanasonicFrame(data []uint32) (int, error) {
	// find start of frame
	for i := 0; i < len(data)-1; i++ {
		if data[i] == (LIRC_MODE2_PULSE|PANASONIC_FRAME_MARK1) && data[i+1] == (LIRC_MODE2_SPACE|PANASONIC_FRAME_MARK2) {
			return i, nil
		}
	}
	return -1, fmt.Errorf("no start of frame found")
}

// Note that the bytes are LittleEndian, and we are reversing the bit order here.
// Values that span two bytes are in LSB order.
func appendBit(space uint32, frame *Frame) error {
	var bit byte
	switch space {
	case PANASONIC_SPACE_0:
		bit = 0
	case PANASONIC_SPACE_1:
		bit = 1
	default:
		return fmt.Errorf("unexpected space length %d", space)
	}
	byteI := frame.NBits / 8
	bitI := frame.NBits % 8 // here is where bit order is reversed
	if byteI == len(frame.Data) {
		frame.Data = append(frame.Data, byte(0))
	}
	frame.Data[byteI] = frame.Data[byteI] | (bit << bitI)
	frame.NBits++
	return nil
}

// Pulses are always the same length, while spaces can be of two lengths that represent 0 and 1.
func parseLircAsPanasonicData(data []uint32, nBits int) (Frame, []uint32, error) {
	start, err := findStartOfPanasonicFrame(data)
	if err != nil {
		return Frame{}, nil, nil
	}
	data = data[start+2:]
	frame := Frame{make([]byte, 0, 16), 0}
	for i := 0; i < len(data); i++ {
		switch data[i] & LIRC_MODE2_MASK {
		case LIRC_MODE2_PULSE:
			if data[i]&LIRC_VALUE_MASK == PANASONIC_PULSE {
				i++
			} else {
				return Frame{}, nil, fmt.Errorf("expected a %d pulse but got data[%d]=%d", PANASONIC_PULSE, i, data[i]&LIRC_VALUE_MASK)
			}
		case LIRC_MODE2_SPACE:
			return Frame{}, nil, fmt.Errorf("expected a pulse but got a space data[%d]=%d", i, data[i]&LIRC_VALUE_MASK)
		default:
			return Frame{}, nil, fmt.Errorf("unexpected LIRC type %08b", data[i]&LIRC_MODE2_MASK)
		}

		// return the frame when all the expected bits have been read
		if frame.NBits == nBits {
			return frame, data[i:], nil
		}

		if i < len(data) {
			switch data[i] & LIRC_MODE2_MASK {
			case LIRC_MODE2_SPACE:
				if data[i]&LIRC_VALUE_MASK == PANASONIC_SEPARATOR {
					// found a frame separator, return current frame and unparsed data
					// fmt.Printf("end-of-frame found %d bits\n", frame.NBits)
					return frame, data[i+1:], nil
				}
				err := appendBit(data[i]&LIRC_VALUE_MASK, &frame)
				if err != nil {
					return Frame{}, nil, err
				}
			case LIRC_MODE2_PULSE:
				return Frame{}, nil, fmt.Errorf("expected a space but got a pulse data[%d]=%d", i, data[i]&LIRC_VALUE_MASK)
			default:
				return Frame{}, nil, fmt.Errorf("unexpected LIRC type %08b", data[i]&LIRC_MODE2_MASK)
			}
		}
	}
	return Frame{}, nil, fmt.Errorf("waiting for all lirc data (including terminating pulse), bits %d < %d", frame.NBits, nBits)
}

func readPanasonicMessage(data []uint32) ([]Frame, []uint32, error) {
	f1, restData, err := parseLircAsPanasonicData(data, PANASONIC_BITS_FRAME1)
	if err != nil {
		return nil, data, err
	}
	f2, restData, err := parseLircAsPanasonicData(restData, PANASONIC_BITS_FRAME2)
	if err != nil {
		return nil, data, err
	}
	return []Frame{f1, f2}, restData, err
}
