package codec

import (
	"fmt"
	"rpi_panasonic_inverter_rc/rcconst"
	"time"
)

type RcConfig struct {
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

	TimerOn  uint
	TimerOff uint

	TimerOnTime  Time
	TimerOffTime Time
	Clock        Time
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

// Return an rcconfig.RcConfig that is initialized from (a received) message.
// If msg is nil, return a default config.
func NewRcConfig() *RcConfig {
	return &RcConfig{
		Power:          rcconst.C_Power_Off,
		Mode:           rcconst.C_Mode_Auto,
		Powerful:       rcconst.C_Powerful_Disabled,
		Quiet:          rcconst.C_Quiet_Disabled,
		Temperature:    20,
		FanSpeed:       rcconst.C_FanSpeed_Auto,
		VentVertical:   rcconst.C_VentVertical_Auto,
		VentHorizontal: rcconst.C_VentHorizontal_Auto,
		TimerOn:        rcconst.C_Timer_Disabled,
		TimerOff:       rcconst.C_Timer_Disabled,
		TimerOnTime:    rcconst.C_Time_Unset,
		TimerOffTime:   rcconst.C_Time_Unset,
		Clock:          rcconst.C_Time_Unset,
	}
}

func RcConfigFromFrame(msg *Message) *RcConfig {
	rcconf := NewRcConfig()

	f := msg.Frame2
	rcconf.Power = f.GetValue(rcconst.P_PANASONIC_POWER_BIT0, rcconst.P_PANASONIC_POWER_BITS)
	rcconf.Mode = f.GetValue(rcconst.P_PANASONIC_MODE_BIT0, rcconst.P_PANASONIC_MODE_BITS)
	rcconf.Powerful = f.GetValue(rcconst.P_PANASONIC_POWERFUL_BIT0, rcconst.P_PANASONIC_POWERFUL_BITS)
	rcconf.Quiet = f.GetValue(rcconst.P_PANASONIC_QUIET_BIT0, rcconst.P_PANASONIC_QUIET_BITS)
	rcconf.Temperature = f.GetValue(rcconst.P_PANASONIC_TEMP_BIT0, rcconst.P_PANASONIC_TEMP_BITS)
	rcconf.FanSpeed = f.GetValue(rcconst.P_PANASONIC_FAN_SPEED_BIT0, rcconst.P_PANASONIC_FAN_SPEED_BITS)
	rcconf.VentVertical = f.GetValue(rcconst.P_PANASONIC_VENT_VPOS_BIT0, rcconst.P_PANASONIC_VENT_VPOS_BITS)
	rcconf.VentHorizontal = f.GetValue(rcconst.P_PANASONIC_VENT_HPOS_BIT0, rcconst.P_PANASONIC_VENT_HPOS_BITS)
	rcconf.TimerOn = f.GetValue(rcconst.P_PANASONIC_TIMER_ON_ENABLED_BIT0, rcconst.P_PANASONIC_TIMER_ON_ENABLED_BITS)
	rcconf.TimerOff = f.GetValue(rcconst.P_PANASONIC_TIMER_OFF_ENABLED_BIT0, rcconst.P_PANASONIC_TIMER_OFF_ENABLED_BITS)
	rcconf.TimerOnTime = Time(f.GetValue(rcconst.P_PANASONIC_TIMER_ON_TIME_BIT0, rcconst.P_PANASONIC_TIMER_ON_TIME_BITS))
	rcconf.TimerOffTime = Time(f.GetValue(rcconst.P_PANASONIC_TIMER_OFF_TIME_BIT0, rcconst.P_PANASONIC_TIMER_OFF_TIME_BITS))
	rcconf.Clock = Time(f.GetValue(rcconst.P_PANASONIC_CLOCK_BIT0, rcconst.P_PANASONIC_CLOCK_BITS))

	return rcconf
}

func (c *RcConfig) CopyForSending() *RcConfig {
	rc := *c
	rc.TimerOnTime = rcconst.C_Time_Unset
	rc.TimerOffTime = rcconst.C_Time_Unset
	rc.Clock = rcconst.C_Time_Unset
	return &rc
}

func (c *RcConfig) SetClock() {
	now := time.Now()
	c.Clock = NewTime(uint(now.Hour()), uint(now.Minute()))
}

func (c *RcConfig) ToMessage() *Message {
	msg := InitializedMessage()
	f := msg.Frame2
	f.SetValue(c.Power, rcconst.P_PANASONIC_POWER_BIT0, rcconst.P_PANASONIC_POWER_BITS).
		SetValue(c.Mode, rcconst.P_PANASONIC_MODE_BIT0, rcconst.P_PANASONIC_MODE_BITS).
		SetValue(c.Powerful, rcconst.P_PANASONIC_POWERFUL_BIT0, rcconst.P_PANASONIC_POWERFUL_BITS).
		SetValue(c.Quiet, rcconst.P_PANASONIC_QUIET_BIT0, rcconst.P_PANASONIC_QUIET_BITS).
		SetValue(c.Temperature, rcconst.P_PANASONIC_TEMP_BIT0, rcconst.P_PANASONIC_TEMP_BITS).
		SetValue(c.FanSpeed, rcconst.P_PANASONIC_FAN_SPEED_BIT0, rcconst.P_PANASONIC_FAN_SPEED_BITS).
		SetValue(c.VentVertical, rcconst.P_PANASONIC_VENT_VPOS_BIT0, rcconst.P_PANASONIC_VENT_VPOS_BITS).
		SetValue(c.VentHorizontal, rcconst.P_PANASONIC_VENT_HPOS_BIT0, rcconst.P_PANASONIC_VENT_HPOS_BITS).
		SetValue(c.TimerOn, rcconst.P_PANASONIC_TIMER_ON_ENABLED_BIT0, rcconst.P_PANASONIC_TIMER_ON_ENABLED_BITS).
		SetValue(c.TimerOff, rcconst.P_PANASONIC_TIMER_OFF_ENABLED_BIT0, rcconst.P_PANASONIC_TIMER_OFF_ENABLED_BITS).
		SetValue(c.TimerOnTime.Minutes(), rcconst.P_PANASONIC_TIMER_ON_TIME_BIT0, rcconst.P_PANASONIC_TIMER_ON_TIME_BITS).
		SetValue(c.TimerOffTime.Minutes(), rcconst.P_PANASONIC_TIMER_OFF_TIME_BIT0, rcconst.P_PANASONIC_TIMER_OFF_TIME_BITS).
		SetValue(c.Clock.Minutes(), rcconst.P_PANASONIC_CLOCK_BIT0, rcconst.P_PANASONIC_CLOCK_BITS)
	return msg
}

func (c *RcConfig) ConvertToLircData() *LircBuffer {
	m := c.ToMessage()
	m.Frame2.SetChecksum()
	return m.ToLirc()
}

func (c *RcConfig) ConvertToMode2LircData() []string {
	m := c.ToMessage()
	m.Frame2.SetChecksum()
	lircData := m.ToLirc()
	s := make([]string, 0, 500)
	for _, v := range lircData.buf {
		if v&rcconst.L_LIRC_MODE2_MASK == rcconst.L_LIRC_MODE2_PULSE {
			s = append(s, fmt.Sprintf("+%d", v&rcconst.L_LIRC_VALUE_MASK))
		} else if v&rcconst.L_LIRC_MODE2_MASK == rcconst.L_LIRC_MODE2_SPACE {
			s = append(s, fmt.Sprintf("-%d", v&rcconst.L_LIRC_VALUE_MASK))
		}
	}
	return s
}
