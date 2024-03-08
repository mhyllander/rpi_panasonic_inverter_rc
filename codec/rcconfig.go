package codec

import (
	"fmt"
	"log/slog"
	"rpi_panasonic_inverter_rc/common"
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

func (t Time) ToString() string {
	return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
}

// Return an rcconfig.RcConfig that is initialized from (a received) message.
// If msg is nil, return a default config.
func NewRcConfig() *RcConfig {
	return &RcConfig{
		Power:          common.C_Power_Off,
		Mode:           common.C_Mode_Auto,
		Powerful:       common.C_Powerful_Disabled,
		Quiet:          common.C_Quiet_Disabled,
		Temperature:    20,
		FanSpeed:       common.C_FanSpeed_Auto,
		VentVertical:   common.C_VentVertical_Auto,
		VentHorizontal: common.C_VentHorizontal_Auto,
		TimerOn:        common.C_Timer_Disabled,
		TimerOff:       common.C_Timer_Disabled,
		TimerOnTime:    common.C_Time_Unset,
		TimerOffTime:   common.C_Time_Unset,
		Clock:          common.C_Time_Unset,
	}
}

func RcConfigFromFrame(msg *Message) *RcConfig {
	rcconf := NewRcConfig()

	f := msg.Frame2
	rcconf.Power = f.GetValue(common.P_PANASONIC_POWER_BIT0, common.P_PANASONIC_POWER_BITS)
	rcconf.Mode = f.GetValue(common.P_PANASONIC_MODE_BIT0, common.P_PANASONIC_MODE_BITS)
	rcconf.Powerful = f.GetValue(common.P_PANASONIC_POWERFUL_BIT0, common.P_PANASONIC_POWERFUL_BITS)
	rcconf.Quiet = f.GetValue(common.P_PANASONIC_QUIET_BIT0, common.P_PANASONIC_QUIET_BITS)
	rcconf.Temperature = f.GetValue(common.P_PANASONIC_TEMP_BIT0, common.P_PANASONIC_TEMP_BITS)
	rcconf.FanSpeed = f.GetValue(common.P_PANASONIC_FAN_SPEED_BIT0, common.P_PANASONIC_FAN_SPEED_BITS)
	rcconf.VentVertical = f.GetValue(common.P_PANASONIC_VENT_VPOS_BIT0, common.P_PANASONIC_VENT_VPOS_BITS)
	rcconf.VentHorizontal = f.GetValue(common.P_PANASONIC_VENT_HPOS_BIT0, common.P_PANASONIC_VENT_HPOS_BITS)
	rcconf.TimerOn = f.GetValue(common.P_PANASONIC_TIMER_ON_ENABLED_BIT0, common.P_PANASONIC_TIMER_ON_ENABLED_BITS)
	rcconf.TimerOff = f.GetValue(common.P_PANASONIC_TIMER_OFF_ENABLED_BIT0, common.P_PANASONIC_TIMER_OFF_ENABLED_BITS)
	rcconf.TimerOnTime = Time(f.GetValue(common.P_PANASONIC_TIMER_ON_TIME_BIT0, common.P_PANASONIC_TIMER_ON_TIME_BITS))
	rcconf.TimerOffTime = Time(f.GetValue(common.P_PANASONIC_TIMER_OFF_TIME_BIT0, common.P_PANASONIC_TIMER_OFF_TIME_BITS))
	rcconf.Clock = Time(f.GetValue(common.P_PANASONIC_CLOCK_BIT0, common.P_PANASONIC_CLOCK_BITS))

	return rcconf
}

// Return the current config with timer times and clock unset, useful
// for sending the updated config to the inverter.
func (c *RcConfig) CopyForSending() *RcConfig {
	rc := *c
	rc.TimerOnTime = common.C_Time_Unset
	rc.TimerOffTime = common.C_Time_Unset
	rc.Clock = common.C_Time_Unset
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
	f.SetValue(c.Power, common.P_PANASONIC_POWER_BIT0, common.P_PANASONIC_POWER_BITS).
		SetValue(c.Mode, common.P_PANASONIC_MODE_BIT0, common.P_PANASONIC_MODE_BITS).
		SetValue(c.Powerful, common.P_PANASONIC_POWERFUL_BIT0, common.P_PANASONIC_POWERFUL_BITS).
		SetValue(c.Quiet, common.P_PANASONIC_QUIET_BIT0, common.P_PANASONIC_QUIET_BITS).
		SetValue(c.Temperature, common.P_PANASONIC_TEMP_BIT0, common.P_PANASONIC_TEMP_BITS).
		SetValue(c.FanSpeed, common.P_PANASONIC_FAN_SPEED_BIT0, common.P_PANASONIC_FAN_SPEED_BITS).
		SetValue(c.VentVertical, common.P_PANASONIC_VENT_VPOS_BIT0, common.P_PANASONIC_VENT_VPOS_BITS).
		SetValue(c.VentHorizontal, common.P_PANASONIC_VENT_HPOS_BIT0, common.P_PANASONIC_VENT_HPOS_BITS).
		SetValue(c.TimerOn, common.P_PANASONIC_TIMER_ON_ENABLED_BIT0, common.P_PANASONIC_TIMER_ON_ENABLED_BITS).
		SetValue(c.TimerOff, common.P_PANASONIC_TIMER_OFF_ENABLED_BIT0, common.P_PANASONIC_TIMER_OFF_ENABLED_BITS).
		SetValue(c.TimerOnTime.Minutes(), common.P_PANASONIC_TIMER_ON_TIME_BIT0, common.P_PANASONIC_TIMER_ON_TIME_BITS).
		SetValue(c.TimerOffTime.Minutes(), common.P_PANASONIC_TIMER_OFF_TIME_BIT0, common.P_PANASONIC_TIMER_OFF_TIME_BITS).
		SetValue(c.Clock.Minutes(), common.P_PANASONIC_CLOCK_BIT0, common.P_PANASONIC_CLOCK_BITS)
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
		if v&common.L_LIRC_MODE2_MASK == common.L_LIRC_MODE2_PULSE {
			s = append(s, fmt.Sprintf("+%d", v&common.L_LIRC_VALUE_MASK))
		} else if v&common.L_LIRC_MODE2_MASK == common.L_LIRC_MODE2_SPACE {
			s = append(s, fmt.Sprintf("-%d", v&common.L_LIRC_VALUE_MASK))
		}
	}
	return s
}

func (c *RcConfig) PrintConfigAndChecksum(checksumStatus string) {
	fmt.Printf("Settings : power=%s(%d) mode=%s(%d) powerful=%s(%d) quiet=%s(%d) temp=%s\n",
		common.Power2String(c.Power), c.Power,
		common.Mode2String(c.Mode), c.Mode,
		common.Powerful2String(c.Powerful), c.Powerful,
		common.Quiet2String(c.Quiet), c.Quiet,
		common.Temperatur2String(c.Temperature))

	fmt.Printf("Air vents: fan=%s(%d) vert=%s(%d) horiz=%s(%d)\n",
		common.FanSpeed2String(c.FanSpeed), c.FanSpeed,
		common.VentVertical2String(c.VentVertical), c.VentVertical,
		common.VentHorizontal2String(c.VentHorizontal), c.VentHorizontal)

	fmt.Printf("Timers   : ton=%s(%d) tont=%s; toff=%s(%d) tofft=%s; clock=%s\n",
		common.TimerToString(c.TimerOn), c.TimerOn, c.TimerOnTime.ToString(), common.TimerToString(c.TimerOff), c.TimerOff, c.TimerOffTime.ToString(), c.Clock.ToString())

	if checksumStatus != "" {
		fmt.Printf("Checksum: %s\n", checksumStatus)
	}
}

func (c *RcConfig) LogConfigAndChecksum(msg, checksumStatus string) {
	if msg == "" {
		msg = "config"
	}
	slog.Info(msg,
		"power", common.Power2String(c.Power),
		"mode", common.Mode2String(c.Mode),
		"powerful", common.Powerful2String(c.Powerful),
		"quiet", common.Quiet2String(c.Quiet),
		"temp", common.Temperatur2String(c.Temperature),
		"fan", common.FanSpeed2String(c.FanSpeed),
		"vert", common.VentVertical2String(c.VentVertical),
		"horiz", common.VentHorizontal2String(c.VentHorizontal),
		"ton", common.TimerToString(c.TimerOn),
		"tont", c.TimerOnTime.ToString(),
		"toff", common.TimerToString(c.TimerOff),
		"tofft", c.TimerOffTime.ToString(),
		"clock", c.Clock.ToString(),
		"checksum", checksumStatus,
	)
}
