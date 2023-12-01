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

	// mode
	PANASONIC_MODE_BIT0 = 44
	PANASONIC_MODE_BITS = 4
	// values
	PANASONIC_MODE_AUTO = 0
	PANASONIC_MODE_DRY  = 2
	PANASONIC_MODE_COOL = 3
	PANASONIC_MODE_HEAT = 4

	// temperature
	// defaults: the RC remembers the temperature for each mode
	PANASONIC_TEMP_BIT0 = 49
	PANASONIC_TEMP_BITS = 5
	// values
	PANASONIC_TEMP_MIN = 16
	PANASONIC_TEMP_MAX = 30

	// vent vertical position
	PANASONIC_VENT_VPOS_BIT0 = 64
	PANASONIC_VENT_VPOS_BITS = 4
	// values
	PANASONIC_VENT_VPOS_HIGHEST = 1
	PANASONIC_VENT_VPOS_HIGH    = 2
	PANASONIC_VENT_VPOS_MIDDLE  = 3
	PANASONIC_VENT_VPOS_LOW     = 4
	PANASONIC_VENT_VPOS_LOWEST  = 5
	PANASONIC_VENT_VPOS_AUTO    = 15

	// fan speed
	// defaults: the RC remembers the fan speed for each mode
	PANASONIC_FAN_SPEED_BIT0 = 68
	PANASONIC_FAN_SPEED_BITS = 4
	// values
	PANASONIC_FAN_SPEED_LOWEST  = 3
	PANASONIC_FAN_SPEED_LOW     = 4
	PANASONIC_FAN_SPEED_MIDDLE  = 5
	PANASONIC_FAN_SPEED_HIGH    = 6
	PANASONIC_FAN_SPEED_HIGHEST = 7
	PANASONIC_FAN_SPEED_AUTO    = 10

	// vent horizontal position
	PANASONIC_VENT_HPOS_BIT0 = 72
	PANASONIC_VENT_HPOS_BITS = 4
	// values
	PANASONIC_VENT_HPOS_FARLEFT  = 9
	PANASONIC_VENT_HPOS_LEFT     = 10
	PANASONIC_VENT_HPOS_MIDDLE   = 6
	PANASONIC_VENT_HPOS_RIGHT    = 11
	PANASONIC_VENT_HPOS_FARRIGHT = 12
	PANASONIC_VENT_HPOS_AUTO     = 13

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
	// this is the value in big.Int as bytes, which can be used to initialize a message when sending
	return []byte{0b00000110, 0b00000000, 0b00000000, 0b00000000, 0b00000100, 0b11100000, 0b00100000, 0b00000010}
}

// this is a template for the second frame (152 bits) used by the Panasonic IR Controller A75C3115
func PANASONIC_FRAME2() []byte {
	// this is the value in big.Int as bytes, which can be used to initialize a message when sending
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
