package codec

const (
	LIRC_MODE2_SPACE     = 0x00000000
	LIRC_MODE2_PULSE     = 0x01000000
	LIRC_MODE2_FREQUENCY = 0x02000000
	LIRC_MODE2_TIMEOUT   = 0x03000000
	LIRC_MODE2_OVERFLOW  = 0x04000000

	LIRC_VALUE_MASK = 0x00FFFFFF
	LIRC_MODE2_MASK = 0xFF000000

	// ioctl
	LIRC_SET_REC_TIMEOUT_REPORTS = uint(0x40046919)

	PANASONIC_FRAME_MARK1 = 3500
	PANASONIC_FRAME_MARK2 = 1750
	PANASONIC_SEPARATOR   = 10000
	PANASONIC_PULSE       = 435
	PANASONIC_SPACE_0     = 435
	PANASONIC_SPACE_1     = 1300

	// this is used to filter out spaces that are not part of the data
	PANASONIC_SPACE_OUTLIER = 20000

	LIRC_TIMING_SPREAD = 100

	// The Panasonic IR Controller A75C3115 sends two frames of data each time. The first frame
	// never changes, while the second contains the complete configuration.
	PANASONIC_BITS_FRAME1 = 64
	PANASONIC_BITS_FRAME2 = 152

	// Pulses and spaces required to transmit 2 frames. Each frame begins with two markers (pulse + space),
	// followed by the frame data which starts and ends with a pulse. The frames are separated by a space marker.
	PANASONIC_LIRC_ITEMS = (2 + PANASONIC_BITS_FRAME1*2 + 1) + 1 + (2 + PANASONIC_BITS_FRAME2*2 + 1)

	// power
	PANASONIC_POWER_BYTE = 5
	PANASONIC_POWER_MASK = 0b00000001

	// timer on
	PANASONIC_TIMER_ON_ENABLED_BYTE = 5
	PANASONIC_TIMER_ON_ENABLED_MASK = 0b00000010

	// timer off
	PANASONIC_TIMER_OFF_ENABLED_BYTE = 5
	PANASONIC_TIMER_OFF_ENABLED_MASK = 0b00000100

	// mode: auto=0 heat=4 cool=3 dry=2
	PANASONIC_MODE_BYTE = 5
	PANASONIC_MODE_MASK = 0b11110000

	// temp: 16-30
	// defaults: the RC remembers the temperature for each mode
	PANASONIC_TEMP_BYTE = 6
	PANASONIC_TEMP_MASK = 0b00111110

	// vent vertical position: highest=1 high=2 middle=3 low=4 lowest=5 auto=15
	PANASONIC_VENT_VPOS_BYTE = 8
	PANASONIC_VENT_VPOS_MASK = 0b00001111

	// fan speed: one=3 two=4 three=5 four=6 five=7 auto=10
	// defaults: the RC remembers the fan speed for each mode
	PANASONIC_FAN_BYTE = 8
	PANASONIC_FAN_MASK = 0b11110000

	// vent horizontal position: far_left=9 left=10 center=6 right=11 far_right=12 auto=13
	PANASONIC_VENT_HPOS_BYTE = 9
	PANASONIC_VENT_HPOS_MASK = 0b00001111

	// timer on is 11 bits in bytes 10-11 (LSB, MSB)
	PANASONIC_TIMER_ON_TIME_BYTE1 = 10
	PANASONIC_TIMER_ON_TIME_MASK1 = 0b11111111
	PANASONIC_TIMER_ON_TIME_BYTE2 = 11
	PANASONIC_TIMER_ON_TIME_MASK2 = 0b00000111

	// timer off is 11 bits in bytes 11-12 (LSB, MSB)
	PANASONIC_TIMER_OFF_TIME_BYTE1 = 11
	PANASONIC_TIMER_OFF_TIME_MASK1 = 0b11110000
	PANASONIC_TIMER_OFF_TIME_BYTE2 = 12
	PANASONIC_TIMER_OFF_TIME_MASK2 = 0b01111111

	// powerful (mutually exclusive with quiet)
	PANASONIC_POWERFUL_BYTE = 13
	PANASONIC_POWERFUL_MASK = 0b00000001

	// quiet (mutually exclusive with powerful)
	PANASONIC_QUIET_BYTE = 13
	PANASONIC_QUIET_MASK = 0b00100000

	// clock is 11 bits in bytes 16-17 (LSB, MSB)
	PANASONIC_CLOCK_BYTE1 = 16
	PANASONIC_CLOCK_MASK1 = 0b11111111
	PANASONIC_CLOCK_BYTE2 = 17
	PANASONIC_CLOCK_MASK2 = 0b00000111

	// This value means that a time field is not set. Times are only set when changing a timer,
	// otherwise all times are set to this value and ignored by the unit.
	PANASONIC_TIME_UNSET = 0x600
)

// these are the timings used by the Panasonic IR Controller A75C3115
func PANASONIC_IR_TIMINGS() []uint32 {
	return []uint32{PANASONIC_FRAME_MARK1, PANASONIC_FRAME_MARK2, PANASONIC_SEPARATOR, PANASONIC_PULSE, PANASONIC_SPACE_0, PANASONIC_SPACE_1}
}

// this is the constant first frame (64 bits) used by the Panasonic IR Controller A75C3115
func PANASONIC_FRAME1() []byte {
	return []byte{0b00000010, 0b00100000, 0b11100000, 0b00000100, 0b00000000, 0b00000000, 0b00000000, 0b00000110}
}

// this is a template for the second frame (152 bits) used by the Panasonic IR Controller A75C3115
func PANASONIC_FRAME2() []byte {
	return []byte{
		0b00000010, // 0: unused?
		0b00100000, // 1: unused?
		0b11100000, // 2: unused?
		0b00000100, // 3: unused?
		0b00000000, // 4: unused?
		0b01001000, // 5: power, timer on, timer off (5 unused bits?)
		0b00000000, // 6: unused (2 bits), temperature (5 bits), unused (1 bit)
		0b10000000, // 7: unused?
		0b00000000, // 8: fan speed (4 bits), vent vpos (4 bits)
		0b00000000, // 9: unused (4 bits), vent hpos (4 bits)
		0b00000000, // 10: timer on (8 bits)
		0b00001000, // 11: timer off (4 bits), unused (1 bit), timer on (3 bits)
		0b10000000, // 12: unused (1 bit), timer off (7 bits)
		0b00000000, // 13: powerful, quiet (6 unused bits?)
		0b00000000, // 14: unused?
		0b10000000, // 15: unused?
		0b00000000, // 16: clock time (8 bits)
		0b00000000, // 17: unused (5 bits), clock time (3 bits)
		0b00000000, // 18: checksum (8 bits)
	}
}
