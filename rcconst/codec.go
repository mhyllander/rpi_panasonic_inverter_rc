package rcconst

// ---------------------------------------------------------------------------------------------------
// Constants related to receiving and parsing LIRC
const (
	L_LIRC_MODE2_SPACE     = 0x00000000
	L_LIRC_MODE2_PULSE     = 0x01000000
	L_LIRC_MODE2_FREQUENCY = 0x02000000
	L_LIRC_MODE2_TIMEOUT   = 0x03000000
	L_LIRC_MODE2_OVERFLOW  = 0x04000000

	L_LIRC_VALUE_MASK = 0x00FFFFFF
	L_LIRC_MODE2_MASK = 0xFF000000

	L_PANASONIC_FRAME_MARK1_PULSE = 3500
	L_PANASONIC_FRAME_MARK2_SPACE = 1750
	L_PANASONIC_SEPARATOR         = 10000
	L_PANASONIC_PULSE             = 435
	L_PANASONIC_SPACE_0           = 435
	L_PANASONIC_SPACE_1           = 1300

	// this is used to filter out obviously wrong spaces and pulses
	L_PANASONIC_PULSE_OUTLIER = 4500
	L_PANASONIC_SPACE_OUTLIER = 11000

	L_PANASONIC_TIMING_SPREAD = 200

	// The Panasonic IR Controller A75C3115 sends two frames of data each time. The first frame
	// never changes, while the second contains the complete configuration.
	L_PANASONIC_BITS_FRAME1 = 64
	L_PANASONIC_BITS_FRAME2 = 152

	// Pulses and spaces required to transmit 2 frames. Each frame begins with two markers (pulse + space),
	// followed by the frame data which starts and ends with a pulse. The frames are separated by a space marker.
	L_PANASONIC_LIRC_ITEMS = (2 + L_PANASONIC_BITS_FRAME1*2 + 1) + 1 + (2 + L_PANASONIC_BITS_FRAME2*2 + 1)
)

// these are the timings used by the Panasonic IR Controller A75C3115
func L_PANASONIC_IR_SPACE_TIMINGS() []uint32 {
	return []uint32{L_PANASONIC_FRAME_MARK2_SPACE, L_PANASONIC_SEPARATOR, L_PANASONIC_SPACE_0, L_PANASONIC_SPACE_1}
}
func L_PANASONIC_IR_PULSE_TIMINGS() []uint32 {
	return []uint32{L_PANASONIC_FRAME_MARK1_PULSE, L_PANASONIC_PULSE}
}

// ---------------------------------------------------------------------------------------------------
// Constants related to the Panasonic IR Controller A75C3115 configuration data structure
const (
	// power
	P_PANASONIC_POWER_BIT0 = 40
	P_PANASONIC_POWER_BITS = 1

	// timer on
	P_PANASONIC_TIMER_ON_ENABLED_BIT0 = 41
	P_PANASONIC_TIMER_ON_ENABLED_BITS = 1

	// timer off
	P_PANASONIC_TIMER_OFF_ENABLED_BIT0 = 42
	P_PANASONIC_TIMER_OFF_ENABLED_BITS = 1

	// mode
	P_PANASONIC_MODE_BIT0 = 44
	P_PANASONIC_MODE_BITS = 4
	// values
	P_PANASONIC_MODE_AUTO = 0
	P_PANASONIC_MODE_DRY  = 2
	P_PANASONIC_MODE_COOL = 3
	P_PANASONIC_MODE_HEAT = 4

	// temperature
	// defaults: the RC remembers the temperature for each mode
	P_PANASONIC_TEMP_BIT0 = 49
	P_PANASONIC_TEMP_BITS = 5
	// values
	P_PANASONIC_TEMP_MIN = 16
	P_PANASONIC_TEMP_MAX = 30

	// vent vertical position
	P_PANASONIC_VENT_VPOS_BIT0 = 64
	P_PANASONIC_VENT_VPOS_BITS = 4
	// values
	P_PANASONIC_VENT_VPOS_HIGHEST = 1
	P_PANASONIC_VENT_VPOS_HIGH    = 2
	P_PANASONIC_VENT_VPOS_MIDDLE  = 3
	P_PANASONIC_VENT_VPOS_LOW     = 4
	P_PANASONIC_VENT_VPOS_LOWEST  = 5
	P_PANASONIC_VENT_VPOS_AUTO    = 15

	// fan speed
	// defaults: the RC remembers the fan speed for each mode
	P_PANASONIC_FAN_SPEED_BIT0 = 68
	P_PANASONIC_FAN_SPEED_BITS = 4
	// values
	P_PANASONIC_FAN_SPEED_LOWEST  = 3
	P_PANASONIC_FAN_SPEED_LOW     = 4
	P_PANASONIC_FAN_SPEED_MIDDLE  = 5
	P_PANASONIC_FAN_SPEED_HIGH    = 6
	P_PANASONIC_FAN_SPEED_HIGHEST = 7
	P_PANASONIC_FAN_SPEED_AUTO    = 10

	// vent horizontal position
	P_PANASONIC_VENT_HPOS_BIT0 = 72
	P_PANASONIC_VENT_HPOS_BITS = 4
	// values
	P_PANASONIC_VENT_HPOS_FARLEFT  = 9
	P_PANASONIC_VENT_HPOS_LEFT     = 10
	P_PANASONIC_VENT_HPOS_MIDDLE   = 6
	P_PANASONIC_VENT_HPOS_RIGHT    = 11
	P_PANASONIC_VENT_HPOS_FARRIGHT = 12
	P_PANASONIC_VENT_HPOS_AUTO     = 13

	// timer on is 11 bits
	P_PANASONIC_TIMER_ON_TIME_BIT0 = 80
	P_PANASONIC_TIMER_ON_TIME_BITS = 11

	// timer off is 11 bits
	P_PANASONIC_TIMER_OFF_TIME_BIT0 = 92
	P_PANASONIC_TIMER_OFF_TIME_BITS = 11

	// powerful (mutually exclusive with quiet, overrides fan speed to auto)
	P_PANASONIC_POWERFUL_BIT0 = 104
	P_PANASONIC_POWERFUL_BITS = 1

	// quiet (mutually exclusive with powerful, overrides fan speed to lowest)
	P_PANASONIC_QUIET_BIT0 = 109
	P_PANASONIC_QUIET_BITS = 1

	// clock is 11 bits
	P_PANASONIC_CLOCK_BIT0 = 128
	P_PANASONIC_CLOCK_BITS = 11

	// checksum
	P_PANASONIC_CHECKSUM_BITS = 8

	// generic values for single bit disabled/enabled
	P_PANASONIC_DISABLED = 0
	P_PANASONIC_ENABLED  = 1

	// This value means that a time field is not set. Times are only set when changing a timer,
	// otherwise all times are set to this value and ignored by the unit.
	P_PANASONIC_TIME_UNSET = 0x600
)

// this is the constant first frame (64 bits) used by the Panasonic IR Controller A75C3115
func P_PANASONIC_FRAME1() []byte {
	// this is the value in big.Int as bytes, which can be used to initialize a message when sending
	return []byte{0b00000110, 0b00000000, 0b00000000, 0b00000000, 0b00000100, 0b11100000, 0b00100000, 0b00000010}
}

// this is a template for the second frame (152 bits) used by the Panasonic IR Controller A75C3115
func P_PANASONIC_FRAME2() []byte {
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
