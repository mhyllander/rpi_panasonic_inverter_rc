package codec

// ---------------------------------------------------------------------------------------------------
// Constants related to receiving and parsing LIRC
const (
	l_LIRC_MODE2_SPACE     = 0x00000000
	l_LIRC_MODE2_PULSE     = 0x01000000
	l_LIRC_MODE2_FREQUENCY = 0x02000000
	l_LIRC_MODE2_TIMEOUT   = 0x03000000
	l_LIRC_MODE2_OVERFLOW  = 0x04000000

	l_LIRC_VALUE_MASK = 0x00FFFFFF
	l_LIRC_MODE2_MASK = 0xFF000000

	l_PANASONIC_FRAME_MARK1_PULSE = 3500
	l_PANASONIC_FRAME_MARK2_SPACE = 1750
	l_PANASONIC_SEPARATOR         = 10000
	l_PANASONIC_PULSE             = 435
	l_PANASONIC_SPACE_0           = 435
	l_PANASONIC_SPACE_1           = 1300

	// this is used to filter out obviously wrong spaces and pulses
	l_PANASONIC_PULSE_OUTLIER = 4500
	l_PANASONIC_SPACE_OUTLIER = 11000

	l_PANASONIC_TIMING_SPREAD = 200

	// The Panasonic IR Controller A75C3115 sends two frames of data each time. The first frame
	// never changes, while the second contains the complete configuration.
	l_PANASONIC_BITS_FRAME1 = 64
	l_PANASONIC_BITS_FRAME2 = 152

	// Pulses and spaces required to transmit 2 frames. Each frame begins with two markers (pulse + space),
	// followed by the frame data which starts and ends with a pulse. The frames are separated by a space marker.
	l_PANASONIC_LIRC_ITEMS = (2 + l_PANASONIC_BITS_FRAME1*2 + 1) + 1 + (2 + l_PANASONIC_BITS_FRAME2*2 + 1)
)

// these are the timings used by the Panasonic IR Controller A75C3115
func l_PANASONIC_IR_SPACE_TIMINGS() []uint32 {
	return []uint32{l_PANASONIC_FRAME_MARK2_SPACE, l_PANASONIC_SEPARATOR, l_PANASONIC_SPACE_0, l_PANASONIC_SPACE_1}
}
func l_PANASONIC_IR_PULSE_TIMINGS() []uint32 {
	return []uint32{l_PANASONIC_FRAME_MARK1_PULSE, l_PANASONIC_PULSE}
}

// ---------------------------------------------------------------------------------------------------
// Constants related to the Panasonic IR Controller A75C3115 configuration data structure
const (
	// power
	p_PANASONIC_POWER_BIT0 = 40
	p_PANASONIC_POWER_BITS = 1

	// timer on
	p_PANASONIC_TIMER_ON_ENABLED_BIT0 = 41
	p_PANASONIC_TIMER_ON_ENABLED_BITS = 1

	// timer off
	p_PANASONIC_TIMER_OFF_ENABLED_BIT0 = 42
	p_PANASONIC_TIMER_OFF_ENABLED_BITS = 1

	// mode
	p_PANASONIC_MODE_BIT0 = 44
	p_PANASONIC_MODE_BITS = 4
	// values
	p_PANASONIC_MODE_AUTO = 0
	p_PANASONIC_MODE_DRY  = 2
	p_PANASONIC_MODE_COOL = 3
	p_PANASONIC_MODE_HEAT = 4

	// temperature
	// defaults: the RC remembers the temperature for each mode
	p_PANASONIC_TEMP_BIT0 = 49
	p_PANASONIC_TEMP_BITS = 5

	// vent vertical position
	p_PANASONIC_VENT_VPOS_BIT0 = 64
	p_PANASONIC_VENT_VPOS_BITS = 4
	// values
	p_PANASONIC_VENT_VPOS_HIGHEST = 1
	p_PANASONIC_VENT_VPOS_HIGH    = 2
	p_PANASONIC_VENT_VPOS_MIDDLE  = 3
	p_PANASONIC_VENT_VPOS_LOW     = 4
	p_PANASONIC_VENT_VPOS_LOWEST  = 5
	p_PANASONIC_VENT_VPOS_AUTO    = 15

	// fan speed
	// defaults: the RC remembers the fan speed for each mode
	p_PANASONIC_FAN_SPEED_BIT0 = 68
	p_PANASONIC_FAN_SPEED_BITS = 4
	// values
	p_PANASONIC_FAN_SPEED_LOWEST  = 3
	p_PANASONIC_FAN_SPEED_LOW     = 4
	p_PANASONIC_FAN_SPEED_MIDDLE  = 5
	p_PANASONIC_FAN_SPEED_HIGH    = 6
	p_PANASONIC_FAN_SPEED_HIGHEST = 7
	p_PANASONIC_FAN_SPEED_AUTO    = 10

	// vent horizontal position
	p_PANASONIC_VENT_HPOS_BIT0 = 72
	p_PANASONIC_VENT_HPOS_BITS = 4
	// values
	p_PANASONIC_VENT_HPOS_FARLEFT  = 9
	p_PANASONIC_VENT_HPOS_LEFT     = 10
	p_PANASONIC_VENT_HPOS_MIDDLE   = 6
	p_PANASONIC_VENT_HPOS_RIGHT    = 11
	p_PANASONIC_VENT_HPOS_FARRIGHT = 12
	p_PANASONIC_VENT_HPOS_AUTO     = 13

	// timer on is 11 bits
	p_PANASONIC_TIMER_ON_TIME_BIT0 = 80
	p_PANASONIC_TIMER_ON_TIME_BITS = 11

	// timer off is 11 bits
	p_PANASONIC_TIMER_OFF_TIME_BIT0 = 92
	p_PANASONIC_TIMER_OFF_TIME_BITS = 11

	// powerful (mutually exclusive with quiet, overrides fan speed to auto)
	p_PANASONIC_POWERFUL_BIT0 = 104
	p_PANASONIC_POWERFUL_BITS = 1

	// quiet (mutually exclusive with powerful, overrides fan speed to lowest)
	p_PANASONIC_QUIET_BIT0 = 109
	p_PANASONIC_QUIET_BITS = 1

	// clock is 11 bits
	p_PANASONIC_CLOCK_BIT0 = 128
	p_PANASONIC_CLOCK_BITS = 11

	// checksum
	p_PANASONIC_CHECKSUM_BITS = 8

	// generic values for single bit disabled/enabled
	p_PANASONIC_DISABLED = 0
	p_PANASONIC_ENABLED  = 1

	// This value means that a time field is not set. Times are only set when changing a timer,
	// otherwise all times are set to this value and ignored by the unit.
	p_PANASONIC_TIME_UNSET = 0x600
)

// this is the constant first frame (64 bits) used by the Panasonic IR Controller A75C3115
func p_PANASONIC_FRAME1() []byte {
	// this is the value in big.Int as bytes, which can be used to initialize a message when sending
	return []byte{0b00000110, 0b00000000, 0b00000000, 0b00000000, 0b00000100, 0b11100000, 0b00100000, 0b00000010}
}

// this is a template for the second frame (152 bits) used by the Panasonic IR Controller A75C3115
func p_PANASONIC_FRAME2() []byte {
	// this is the value in big.Int as bytes, which can be used to initialize a message when sending
	return []byte{
		0b00000000, // checksum (8 bits)
		0b00000000, // unused (5 bits), clock time (3 bits)
		0b00000000, // clock time (8 bits)
		0b10000000, // unused?
		0b00000000, // unused?
		0b00000000, // powerful, quiet (6 unused bits?)
		0b10000000, // unused (1 bit), timer off (7 bits)
		0b00001000, // timer off (4 bits), unused (1 bit), timer on (3 bits)
		0b00000000, // timer on (8 bits)
		0b00000000, // unused (4 bits), vent hpos (4 bits)
		0b00000000, // fan speed (4 bits), vent vpos (4 bits)
		0b10000000, // unused?
		0b00000000, // unused (2 bits), temperature (5 bits), unused (1 bit)
		0b01001000, // power, timer on, timer off (5 unused bits?)
		0b00000000, // unused?
		0b00000100, // unused?
		0b11100000, // unused?
		0b00100000, // unused?
		0b00000010, // unused?
	}
}
