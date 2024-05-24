package codec

import (
	"fmt"
	"log/slog"
	"time"

	"rpi_panasonic_inverter_rc/codecbase"
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

func (t Time) ToString() string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

// Return an rcconfig.RcConfig that is initialized from (a received) message.
// If msg is nil, return a default config.
func NewRcConfig() *RcConfig {
	return &RcConfig{
		Power:          codecbase.C_Power_Off,
		Mode:           codecbase.C_Mode_Auto,
		Powerful:       codecbase.C_Powerful_Disabled,
		Quiet:          codecbase.C_Quiet_Disabled,
		Temperature:    20,
		FanSpeed:       codecbase.C_FanSpeed_Auto,
		VentVertical:   codecbase.C_VentVertical_Auto,
		VentHorizontal: codecbase.C_VentHorizontal_Auto,
		TimerOn:        codecbase.C_Timer_Disabled,
		TimerOff:       codecbase.C_Timer_Disabled,
		TimerOnTime:    codecbase.C_Time_Unset,
		TimerOffTime:   codecbase.C_Time_Unset,
		Clock:          codecbase.C_Time_Unset,
	}
}

func RcConfigFromFrame(msg *Message) *RcConfig {
	rcconf := NewRcConfig()

	f := msg.Frame2
	rcconf.Power = f.GetValue(codecbase.P_PANASONIC_POWER_BIT0, codecbase.P_PANASONIC_POWER_BITS)
	rcconf.Mode = f.GetValue(codecbase.P_PANASONIC_MODE_BIT0, codecbase.P_PANASONIC_MODE_BITS)
	rcconf.Powerful = f.GetValue(codecbase.P_PANASONIC_POWERFUL_BIT0, codecbase.P_PANASONIC_POWERFUL_BITS)
	rcconf.Quiet = f.GetValue(codecbase.P_PANASONIC_QUIET_BIT0, codecbase.P_PANASONIC_QUIET_BITS)
	rcconf.Temperature = f.GetValue(codecbase.P_PANASONIC_TEMP_BIT0, codecbase.P_PANASONIC_TEMP_BITS)
	rcconf.FanSpeed = f.GetValue(codecbase.P_PANASONIC_FAN_SPEED_BIT0, codecbase.P_PANASONIC_FAN_SPEED_BITS)
	rcconf.VentVertical = f.GetValue(codecbase.P_PANASONIC_VENT_VPOS_BIT0, codecbase.P_PANASONIC_VENT_VPOS_BITS)
	rcconf.VentHorizontal = f.GetValue(codecbase.P_PANASONIC_VENT_HPOS_BIT0, codecbase.P_PANASONIC_VENT_HPOS_BITS)
	rcconf.TimerOn = f.GetValue(codecbase.P_PANASONIC_TIMER_ON_ENABLED_BIT0, codecbase.P_PANASONIC_TIMER_ON_ENABLED_BITS)
	rcconf.TimerOff = f.GetValue(codecbase.P_PANASONIC_TIMER_OFF_ENABLED_BIT0, codecbase.P_PANASONIC_TIMER_OFF_ENABLED_BITS)
	rcconf.TimerOnTime = Time(f.GetValue(codecbase.P_PANASONIC_TIMER_ON_TIME_BIT0, codecbase.P_PANASONIC_TIMER_ON_TIME_BITS))
	rcconf.TimerOffTime = Time(f.GetValue(codecbase.P_PANASONIC_TIMER_OFF_TIME_BIT0, codecbase.P_PANASONIC_TIMER_OFF_TIME_BITS))
	rcconf.Clock = Time(f.GetValue(codecbase.P_PANASONIC_CLOCK_BIT0, codecbase.P_PANASONIC_CLOCK_BITS))

	return rcconf
}

// Return the current config with timer times and clock unset, useful
// for sending the updated config to the inverter.
func (c *RcConfig) CopyForSending() *RcConfig {
	rc := *c
	rc.TimerOnTime = codecbase.C_Time_Unset
	rc.TimerOffTime = codecbase.C_Time_Unset
	rc.Clock = codecbase.C_Time_Unset
	return &rc
}

// Return the current config with initialized clock, useful for
// immediately sending the current config to the inverter after e.g.
// a power outage. It could be used to send the current configuration
// after the RPi has booted up.
func (c *RcConfig) CopyForSendingAll() *RcConfig {
	rc := *c
	rc.SetClock()
	return &rc
}

func (c *RcConfig) SetClock() {
	now := time.Now()
	c.Clock = NewTime(uint(now.Hour()), uint(now.Minute()))
}

func (c *RcConfig) ToMessage() *Message {
	msg := InitializedMessage()
	f := msg.Frame2
	f.SetValue(c.Power, codecbase.P_PANASONIC_POWER_BIT0, codecbase.P_PANASONIC_POWER_BITS).
		SetValue(c.Mode, codecbase.P_PANASONIC_MODE_BIT0, codecbase.P_PANASONIC_MODE_BITS).
		SetValue(c.Powerful, codecbase.P_PANASONIC_POWERFUL_BIT0, codecbase.P_PANASONIC_POWERFUL_BITS).
		SetValue(c.Quiet, codecbase.P_PANASONIC_QUIET_BIT0, codecbase.P_PANASONIC_QUIET_BITS).
		SetValue(c.Temperature, codecbase.P_PANASONIC_TEMP_BIT0, codecbase.P_PANASONIC_TEMP_BITS).
		SetValue(c.FanSpeed, codecbase.P_PANASONIC_FAN_SPEED_BIT0, codecbase.P_PANASONIC_FAN_SPEED_BITS).
		SetValue(c.VentVertical, codecbase.P_PANASONIC_VENT_VPOS_BIT0, codecbase.P_PANASONIC_VENT_VPOS_BITS).
		SetValue(c.VentHorizontal, codecbase.P_PANASONIC_VENT_HPOS_BIT0, codecbase.P_PANASONIC_VENT_HPOS_BITS).
		SetValue(c.TimerOn, codecbase.P_PANASONIC_TIMER_ON_ENABLED_BIT0, codecbase.P_PANASONIC_TIMER_ON_ENABLED_BITS).
		SetValue(c.TimerOff, codecbase.P_PANASONIC_TIMER_OFF_ENABLED_BIT0, codecbase.P_PANASONIC_TIMER_OFF_ENABLED_BITS).
		SetValue(c.TimerOnTime.Minutes(), codecbase.P_PANASONIC_TIMER_ON_TIME_BIT0, codecbase.P_PANASONIC_TIMER_ON_TIME_BITS).
		SetValue(c.TimerOffTime.Minutes(), codecbase.P_PANASONIC_TIMER_OFF_TIME_BIT0, codecbase.P_PANASONIC_TIMER_OFF_TIME_BITS).
		SetValue(c.Clock.Minutes(), codecbase.P_PANASONIC_CLOCK_BIT0, codecbase.P_PANASONIC_CLOCK_BITS)
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
		if v&codecbase.L_LIRC_MODE2_MASK == codecbase.L_LIRC_MODE2_PULSE {
			s = append(s, fmt.Sprintf("+%d", v&codecbase.L_LIRC_VALUE_MASK))
		} else if v&codecbase.L_LIRC_MODE2_MASK == codecbase.L_LIRC_MODE2_SPACE {
			s = append(s, fmt.Sprintf("-%d", v&codecbase.L_LIRC_VALUE_MASK))
		}
	}
	return s
}

func (c *RcConfig) PrintConfigAndChecksum(checksumStatus string) {
	fmt.Printf("Settings : power=%s(%d) mode=%s(%d) powerful=%s(%d) quiet=%s(%d) temp=%s\n",
		codecbase.Power2String(c.Power), c.Power,
		codecbase.Mode2String(c.Mode), c.Mode,
		codecbase.Powerful2String(c.Powerful), c.Powerful,
		codecbase.Quiet2String(c.Quiet), c.Quiet,
		codecbase.Temperatur2String(c.Temperature))

	fmt.Printf("Air vents: fan=%s(%d) vert=%s(%d) horiz=%s(%d)\n",
		codecbase.FanSpeed2String(c.FanSpeed), c.FanSpeed,
		codecbase.VentVertical2String(c.VentVertical), c.VentVertical,
		codecbase.VentHorizontal2String(c.VentHorizontal), c.VentHorizontal)

	fmt.Printf("Timers   : ton=%s(%d) tont=%s; toff=%s(%d) tofft=%s; clock=%s\n",
		codecbase.TimerToString(c.TimerOn), c.TimerOn, c.TimerOnTime.ToString(), codecbase.TimerToString(c.TimerOff), c.TimerOff, c.TimerOffTime.ToString(), c.Clock.ToString())

	if checksumStatus != "" {
		fmt.Printf("Checksum: %s\n", checksumStatus)
	}
}

func (c *RcConfig) LogConfigAndChecksum(msg, checksumStatus string) {
	if msg == "" {
		msg = "config"
	}
	slog.Info(msg,
		"power", codecbase.Power2String(c.Power),
		"mode", codecbase.Mode2String(c.Mode),
		"powerful", codecbase.Powerful2String(c.Powerful),
		"quiet", codecbase.Quiet2String(c.Quiet),
		"temp", codecbase.Temperatur2String(c.Temperature),
		"fan", codecbase.FanSpeed2String(c.FanSpeed),
		"vert", codecbase.VentVertical2String(c.VentVertical),
		"horiz", codecbase.VentHorizontal2String(c.VentHorizontal),
		"ton", codecbase.TimerToString(c.TimerOn),
		"tont", c.TimerOnTime.ToString(),
		"toff", codecbase.TimerToString(c.TimerOff),
		"tofft", c.TimerOffTime.ToString(),
		"clock", c.Clock.ToString(),
		"checksum", checksumStatus,
	)
}
