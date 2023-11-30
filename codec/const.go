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
	PANASONIC_POWER_BIT0 = 40
	PANASONIC_POWER_BITS = 1

	// timer on
	PANASONIC_TIMER_ON_ENABLED_BIT0 = 41
	PANASONIC_TIMER_ON_ENABLED_BITS = 1

	// timer off
	PANASONIC_TIMER_OFF_ENABLED_BIT0 = 42
	PANASONIC_TIMER_OFF_ENABLED_BITS = 1

	// mode: auto=0 heat=4 cool=3 dry=2
	PANASONIC_MODE_BIT0 = 44
	PANASONIC_MODE_BITS = 4

	// temp: 16-30
	// defaults: the RC remembers the temperature for each mode
	PANASONIC_TEMP_BIT0 = 49
	PANASONIC_TEMP_BITS = 5

	// vent vertical position: highest=1 high=2 middle=3 low=4 lowest=5 auto=15
	PANASONIC_VENT_VPOS_BIT0 = 64
	PANASONIC_VENT_VPOS_BITS = 4

	// fan speed: one=3 two=4 three=5 four=6 five=7 auto=10
	// defaults: the RC remembers the fan speed for each mode
	PANASONIC_FAN_BIT0 = 68
	PANASONIC_FAN_BITS = 4

	// vent horizontal position: far_left=9 left=10 center=6 right=11 far_right=12 auto=13
	PANASONIC_VENT_HPOS_BIT0 = 72
	PANASONIC_VENT_HPOS_BITS = 4

	// timer on is 11 bits
	PANASONIC_TIMER_ON_TIME_BIT0 = 80
	PANASONIC_TIMER_ON_TIME_BITS = 11

	// timer off is 11 bits
	PANASONIC_TIMER_OFF_TIME_BIT0 = 92
	PANASONIC_TIMER_OFF_TIME_BITS = 11

	// powerful (mutually exclusive with quiet, sets fan speed to auto)
	PANASONIC_POWERFUL_BIT0 = 104
	PANASONIC_POWERFUL_BITS = 1

	// quiet (mutually exclusive with powerful, sets fan speed to one=3)
	PANASONIC_QUIET_BIT0 = 109
	PANASONIC_QUIET_BITS = 1

	// clock is 11 bits
	PANASONIC_CLOCK_BIT0 = 128
	PANASONIC_CLOCK_BITS = 11

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
	// this is the value in big.Int, shown as bytes, which can be used to initialize the frames when sending
	return []byte{0b00000110, 0b00000000, 0b00000000, 0b00000000, 0b00000100, 0b11100000, 0b00100000, 0b00000010}
}

// this is a template for the second frame (152 bits) used by the Panasonic IR Controller A75C3115
func PANASONIC_FRAME2() []byte {
	// this is the value in big.Int, shown as bytes, which can be used to initialize the frames when sending
	return []byte{
		0b00000000, // 18: checksum (8 bits)
		0b00000000, // 17: unused (5 bits), clock time (3 bits)
		0b00000000, // 16: clock time (8 bits)
		0b10000000, // 15: unused?
		0b00000000, // 14: unused?
		0b00000000, // 13: powerful, quiet (6 unused bits?)
		0b10000000, // 12: unused (1 bit), timer off (7 bits)
		0b00001000, // 11: timer off (4 bits), unused (1 bit), timer on (3 bits)
		0b00000000, // 10: timer on (8 bits)
		0b00000000, // 9: unused (4 bits), vent hpos (4 bits)
		0b00000000, // 8: fan speed (4 bits), vent vpos (4 bits)
		0b10000000, // 7: unused?
		0b00000000, // 6: unused (2 bits), temperature (5 bits), unused (1 bit)
		0b01001000, // 5: power, timer on, timer off (5 unused bits?)
		0b00000000, // 4: unused?
		0b00000100, // 3: unused?
		0b11100000, // 2: unused?
		0b00100000, // 1: unused?
		0b00000010, // 0: unused?
	}
}
