package codec

import "fmt"

const (
	Power_Off = PANASONIC_DISABLED
	Power_On  = PANASONIC_ENABLED

	Mode_Auto = PANASONIC_MODE_AUTO
	Mode_Dry  = PANASONIC_MODE_DRY
	Mode_Cool = PANASONIC_MODE_COOL
	Mode_Heat = PANASONIC_MODE_HEAT

	Powerful_Disabled = PANASONIC_DISABLED
	Powerful_Enabled  = PANASONIC_ENABLED

	Quiet_Disabled = PANASONIC_DISABLED
	Quiet_Enabled  = PANASONIC_ENABLED

	FanSpeed_Auto    = PANASONIC_FAN_SPEED_AUTO
	FanSpeed_Lowest  = PANASONIC_FAN_SPEED_LOWEST
	FanSpeed_Low     = PANASONIC_FAN_SPEED_LOW
	FanSpeed_Middle  = PANASONIC_FAN_SPEED_MIDDLE
	FanSpeed_High    = PANASONIC_FAN_SPEED_HIGH
	FanSpeed_Highest = PANASONIC_FAN_SPEED_HIGHEST

	VentVertical_Auto    = PANASONIC_VENT_VPOS_AUTO
	VentVertical_Lowest  = PANASONIC_VENT_VPOS_LOWEST
	VentVertical_Low     = PANASONIC_VENT_VPOS_LOW
	VentVertical_Middle  = PANASONIC_VENT_VPOS_MIDDLE
	VentVertical_High    = PANASONIC_VENT_VPOS_HIGH
	VentVertical_Highest = PANASONIC_VENT_VPOS_HIGHEST

	VentHorizontal_Auto     = PANASONIC_VENT_HPOS_AUTO
	VentHorizontal_FarLeft  = PANASONIC_VENT_HPOS_FARLEFT
	VentHorizontal_Left     = PANASONIC_VENT_HPOS_LEFT
	VentHorizontal_Middle   = PANASONIC_VENT_HPOS_MIDDLE
	VentHorizontal_Right    = PANASONIC_VENT_HPOS_RIGHT
	VentHorizontal_FarRight = PANASONIC_VENT_HPOS_FARRIGHT

	Timer_Disabled = PANASONIC_DISABLED
	Timer_Enabled  = PANASONIC_ENABLED
)

type IrConfig struct {
	Power    uint8
	Mode     uint8
	Powerful uint8
	Quiet    uint8

	// Temperature and FanSpeed depend on Mode, Powerful and Quiet
	Temperature uint8
	FanSpeed    uint8

	// direction of airflow
	VentVertical   uint8
	VentHorizontal uint8

	TimerOnEnabled  uint8
	TimerOffEnabled uint8

	TimerOn  Time
	TimerOff Time
	Clock    Time
}

type Time uint16

func NewTime(hour, minute uint8) Time {
	return Time(uint16(hour)*60 + uint16(minute))
}

func (t Time) Hour() uint8 {
	return uint8(t / 60)
}

func (t Time) Minute() uint8 {
	return uint8(t % 60)
}

func (t Time) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

// Return an IrConfig that is initialized from (a received) message.
// If msg is nil, return a default config.
func NewIrConfig(msg *Message) IrConfig {
	irconf := IrConfig{
		Power:           Power_Off,
		Mode:            Mode_Auto,
		Powerful:        Powerful_Disabled,
		Quiet:           Quiet_Disabled,
		Temperature:     20,
		FanSpeed:        FanSpeed_Auto,
		VentVertical:    VentVertical_Auto,
		VentHorizontal:  VentHorizontal_Auto,
		TimerOnEnabled:  Timer_Disabled,
		TimerOffEnabled: Timer_Disabled,
		TimerOn:         PANASONIC_TIME_UNSET,
		TimerOff:        PANASONIC_TIME_UNSET,
		Clock:           PANASONIC_TIME_UNSET,
	}
	if msg == nil {
		return irconf
	}

	f := msg.Frame2
	irconf.Power = uint8(f.GetValue(PANASONIC_POWER_BIT0, PANASONIC_POWER_BITS))
	irconf.Mode = uint8(f.GetValue(PANASONIC_MODE_BIT0, PANASONIC_MODE_BITS))
	irconf.Powerful = uint8(f.GetValue(PANASONIC_POWERFUL_BIT0, PANASONIC_POWERFUL_BITS))
	irconf.Quiet = uint8(f.GetValue(PANASONIC_QUIET_BIT0, PANASONIC_QUIET_BITS))
	irconf.Temperature = uint8(f.GetValue(PANASONIC_TEMP_BIT0, PANASONIC_TEMP_BITS))
	irconf.FanSpeed = uint8(f.GetValue(PANASONIC_FAN_SPEED_BIT0, PANASONIC_FAN_SPEED_BITS))
	irconf.VentVertical = uint8(f.GetValue(PANASONIC_VENT_VPOS_BIT0, PANASONIC_VENT_VPOS_BITS))
	irconf.VentHorizontal = uint8(f.GetValue(PANASONIC_VENT_HPOS_BIT0, PANASONIC_VENT_HPOS_BITS))
	irconf.TimerOnEnabled = uint8(f.GetValue(PANASONIC_TIMER_ON_ENABLED_BIT0, PANASONIC_TIMER_ON_ENABLED_BITS))
	irconf.TimerOffEnabled = uint8(f.GetValue(PANASONIC_TIMER_OFF_ENABLED_BIT0, PANASONIC_TIMER_OFF_ENABLED_BITS))

	timer_on_time := f.GetValue(PANASONIC_TIMER_ON_TIME_BIT0, PANASONIC_TIMER_ON_TIME_BITS)
	irconf.TimerOn = Time(timer_on_time)
	timer_off_time := f.GetValue(PANASONIC_TIMER_OFF_TIME_BIT0, PANASONIC_TIMER_OFF_TIME_BITS)
	irconf.TimerOff = Time(timer_off_time)
	clock_time := f.GetValue(PANASONIC_CLOCK_BIT0, PANASONIC_CLOCK_BITS)
	irconf.Clock = Time(clock_time)

	return irconf
}

func (c IrConfig) ToMessage() *Message {
	msg := InitializedMessage()
	f := msg.Frame2
	f.SetValue(uint64(c.Power), PANASONIC_POWER_BIT0, PANASONIC_POWER_BITS).
		SetValue(uint64(c.Mode), PANASONIC_MODE_BIT0, PANASONIC_MODE_BITS).
		SetValue(uint64(c.Powerful), PANASONIC_POWERFUL_BIT0, PANASONIC_POWERFUL_BITS).
		SetValue(uint64(c.Quiet), PANASONIC_QUIET_BIT0, PANASONIC_QUIET_BITS).
		SetValue(uint64(c.Temperature), PANASONIC_TEMP_BIT0, PANASONIC_TEMP_BITS).
		SetValue(uint64(c.FanSpeed), PANASONIC_FAN_SPEED_BIT0, PANASONIC_FAN_SPEED_BITS).
		SetValue(uint64(c.VentVertical), PANASONIC_VENT_VPOS_BIT0, PANASONIC_VENT_VPOS_BITS).
		SetValue(uint64(c.VentHorizontal), PANASONIC_VENT_HPOS_BIT0, PANASONIC_VENT_HPOS_BITS).
		SetValue(uint64(c.TimerOnEnabled), PANASONIC_TIMER_ON_ENABLED_BIT0, PANASONIC_TIMER_ON_ENABLED_BITS).
		SetValue(uint64(c.TimerOffEnabled), PANASONIC_TIMER_OFF_ENABLED_BIT0, PANASONIC_TIMER_OFF_ENABLED_BITS).
		SetValue(uint64(c.TimerOn), PANASONIC_TIMER_ON_TIME_BIT0, PANASONIC_TIMER_ON_TIME_BITS).
		SetValue(uint64(c.TimerOff), PANASONIC_TIMER_OFF_TIME_BIT0, PANASONIC_TIMER_OFF_TIME_BITS).
		SetValue(uint64(c.Clock), PANASONIC_CLOCK_BIT0, PANASONIC_CLOCK_BITS)
	return msg
}
