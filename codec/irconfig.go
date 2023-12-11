package codec

import "fmt"

const (
	C_Power_Off = p_PANASONIC_DISABLED
	C_Power_On  = p_PANASONIC_ENABLED

	C_Mode_Auto = p_PANASONIC_MODE_AUTO
	C_Mode_Dry  = p_PANASONIC_MODE_DRY
	C_Mode_Cool = p_PANASONIC_MODE_COOL
	C_Mode_Heat = p_PANASONIC_MODE_HEAT

	C_Powerful_Disabled = p_PANASONIC_DISABLED
	C_Powerful_Enabled  = p_PANASONIC_ENABLED

	C_Quiet_Disabled = p_PANASONIC_DISABLED
	C_Quiet_Enabled  = p_PANASONIC_ENABLED

	C_Temp_Min = 16
	C_Temp_Max = 30

	C_FanSpeed_Auto    = p_PANASONIC_FAN_SPEED_AUTO
	C_FanSpeed_Lowest  = p_PANASONIC_FAN_SPEED_LOWEST
	C_FanSpeed_Low     = p_PANASONIC_FAN_SPEED_LOW
	C_FanSpeed_Middle  = p_PANASONIC_FAN_SPEED_MIDDLE
	C_FanSpeed_High    = p_PANASONIC_FAN_SPEED_HIGH
	C_FanSpeed_Highest = p_PANASONIC_FAN_SPEED_HIGHEST

	C_VentVertical_Auto    = p_PANASONIC_VENT_VPOS_AUTO
	C_VentVertical_Lowest  = p_PANASONIC_VENT_VPOS_LOWEST
	C_VentVertical_Low     = p_PANASONIC_VENT_VPOS_LOW
	C_VentVertical_Middle  = p_PANASONIC_VENT_VPOS_MIDDLE
	C_VentVertical_High    = p_PANASONIC_VENT_VPOS_HIGH
	C_VentVertical_Highest = p_PANASONIC_VENT_VPOS_HIGHEST

	C_VentHorizontal_Auto     = p_PANASONIC_VENT_HPOS_AUTO
	C_VentHorizontal_FarLeft  = p_PANASONIC_VENT_HPOS_FARLEFT
	C_VentHorizontal_Left     = p_PANASONIC_VENT_HPOS_LEFT
	C_VentHorizontal_Middle   = p_PANASONIC_VENT_HPOS_MIDDLE
	C_VentHorizontal_Right    = p_PANASONIC_VENT_HPOS_RIGHT
	C_VentHorizontal_FarRight = p_PANASONIC_VENT_HPOS_FARRIGHT

	C_Timer_Disabled = p_PANASONIC_DISABLED
	C_Timer_Enabled  = p_PANASONIC_ENABLED

	C_Time_Unset = p_PANASONIC_TIME_UNSET
)

type IrConfig struct {
	Power    uint
	Mode     uint
	Powerful uint
	Quiet    uint

	// Temperature is set per Mode
	Temperature uint

	// FanSpeed is set per Mode, but:
	//   Powerful overrides FanSpeed=Auto
	//   Quiet overrides FanSpeed=Lowest
	FanSpeed uint

	// direction of airflow
	VentVertical   uint
	VentHorizontal uint

	TimerOnEnabled  uint
	TimerOffEnabled uint

	TimerOn  Time
	TimerOff Time
	Clock    Time
}

type Time uint

func NewTime(hour, minute uint) Time {
	return Time(hour*60 + minute)
}

func (t Time) Minutes() uint {
	return uint(t)
}

func (t Time) Hour() uint {
	return uint(t) / 60
}

func (t Time) Minute() uint {
	return uint(t) % 60
}

func (t Time) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

// Return an IrConfig that is initialized from (a received) message.
// If msg is nil, return a default config.
func NewIrConfig(msg *Message) *IrConfig {
	irconf := IrConfig{
		Power:           C_Power_Off,
		Mode:            C_Mode_Auto,
		Powerful:        C_Powerful_Disabled,
		Quiet:           C_Quiet_Disabled,
		Temperature:     20,
		FanSpeed:        C_FanSpeed_Auto,
		VentVertical:    C_VentVertical_Auto,
		VentHorizontal:  C_VentHorizontal_Auto,
		TimerOnEnabled:  C_Timer_Disabled,
		TimerOffEnabled: C_Timer_Disabled,
		TimerOn:         C_Time_Unset,
		TimerOff:        C_Time_Unset,
		Clock:           C_Time_Unset,
	}
	if msg == nil {
		return &irconf
	}

	f := msg.Frame2
	irconf.Power = f.GetValue(p_PANASONIC_POWER_BIT0, p_PANASONIC_POWER_BITS)
	irconf.Mode = f.GetValue(p_PANASONIC_MODE_BIT0, p_PANASONIC_MODE_BITS)
	irconf.Powerful = f.GetValue(p_PANASONIC_POWERFUL_BIT0, p_PANASONIC_POWERFUL_BITS)
	irconf.Quiet = f.GetValue(p_PANASONIC_QUIET_BIT0, p_PANASONIC_QUIET_BITS)
	irconf.Temperature = f.GetValue(p_PANASONIC_TEMP_BIT0, p_PANASONIC_TEMP_BITS)
	irconf.FanSpeed = f.GetValue(p_PANASONIC_FAN_SPEED_BIT0, p_PANASONIC_FAN_SPEED_BITS)
	irconf.VentVertical = f.GetValue(p_PANASONIC_VENT_VPOS_BIT0, p_PANASONIC_VENT_VPOS_BITS)
	irconf.VentHorizontal = f.GetValue(p_PANASONIC_VENT_HPOS_BIT0, p_PANASONIC_VENT_HPOS_BITS)
	irconf.TimerOnEnabled = f.GetValue(p_PANASONIC_TIMER_ON_ENABLED_BIT0, p_PANASONIC_TIMER_ON_ENABLED_BITS)
	irconf.TimerOffEnabled = f.GetValue(p_PANASONIC_TIMER_OFF_ENABLED_BIT0, p_PANASONIC_TIMER_OFF_ENABLED_BITS)
	irconf.TimerOn = Time(f.GetValue(p_PANASONIC_TIMER_ON_TIME_BIT0, p_PANASONIC_TIMER_ON_TIME_BITS))
	irconf.TimerOff = Time(f.GetValue(p_PANASONIC_TIMER_OFF_TIME_BIT0, p_PANASONIC_TIMER_OFF_TIME_BITS))
	irconf.Clock = Time(f.GetValue(p_PANASONIC_CLOCK_BIT0, p_PANASONIC_CLOCK_BITS))

	return &irconf
}

func (c *IrConfig) CopyForSending() *IrConfig {
	ic := *c
	ic.TimerOn = C_Time_Unset
	ic.TimerOff = C_Time_Unset
	ic.Clock = C_Time_Unset
	return &ic
}

func (c *IrConfig) ToMessage() *Message {
	msg := InitializedMessage()
	f := msg.Frame2
	f.SetValue(c.Power, p_PANASONIC_POWER_BIT0, p_PANASONIC_POWER_BITS).
		SetValue(c.Mode, p_PANASONIC_MODE_BIT0, p_PANASONIC_MODE_BITS).
		SetValue(c.Powerful, p_PANASONIC_POWERFUL_BIT0, p_PANASONIC_POWERFUL_BITS).
		SetValue(c.Quiet, p_PANASONIC_QUIET_BIT0, p_PANASONIC_QUIET_BITS).
		SetValue(c.Temperature, p_PANASONIC_TEMP_BIT0, p_PANASONIC_TEMP_BITS).
		SetValue(c.FanSpeed, p_PANASONIC_FAN_SPEED_BIT0, p_PANASONIC_FAN_SPEED_BITS).
		SetValue(c.VentVertical, p_PANASONIC_VENT_VPOS_BIT0, p_PANASONIC_VENT_VPOS_BITS).
		SetValue(c.VentHorizontal, p_PANASONIC_VENT_HPOS_BIT0, p_PANASONIC_VENT_HPOS_BITS).
		SetValue(c.TimerOnEnabled, p_PANASONIC_TIMER_ON_ENABLED_BIT0, p_PANASONIC_TIMER_ON_ENABLED_BITS).
		SetValue(c.TimerOffEnabled, p_PANASONIC_TIMER_OFF_ENABLED_BIT0, p_PANASONIC_TIMER_OFF_ENABLED_BITS).
		SetValue(c.TimerOn.Minutes(), p_PANASONIC_TIMER_ON_TIME_BIT0, p_PANASONIC_TIMER_ON_TIME_BITS).
		SetValue(c.TimerOff.Minutes(), p_PANASONIC_TIMER_OFF_TIME_BIT0, p_PANASONIC_TIMER_OFF_TIME_BITS).
		SetValue(c.Clock.Minutes(), p_PANASONIC_CLOCK_BIT0, p_PANASONIC_CLOCK_BITS)
	return msg
}

func (c *IrConfig) ConvertToLircData() *LircBuffer {
	m := c.ToMessage()
	m.Frame2.SetChecksum()
	return m.ToLirc()
}

func (c *IrConfig) ConvertToMode2LircData() []string {
	m := c.ToMessage()
	m.Frame2.SetChecksum()
	lircData := m.ToLirc()
	s := make([]string, 0, 500)
	for _, v := range lircData.buf {
		if v&l_LIRC_MODE2_MASK == l_LIRC_MODE2_PULSE {
			s = append(s, fmt.Sprintf("+%d", v&l_LIRC_VALUE_MASK))
		} else if v&l_LIRC_MODE2_MASK == l_LIRC_MODE2_SPACE {
			s = append(s, fmt.Sprintf("-%d", v&l_LIRC_VALUE_MASK))
		}
	}
	return s
}
