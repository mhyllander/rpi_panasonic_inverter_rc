package utils

import (
	"fmt"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/rcconst"
	"strconv"
	"strings"
	"time"
)

func SetPower(setting string, rc, dbRc *codec.RcConfig) {
	now := time.Now()
	setPower(setting, rc, dbRc, codec.NewTime(uint(now.Hour()), uint(now.Minute())))
}

func setPower(setting string, rc, dbRc *codec.RcConfig, clock codec.Time) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		rc.Power = rcconst.C_Power_On
	case "off", "no", "disable", "disabled":
		rc.Power = rcconst.C_Power_Off
	default:
		// Adjust Power according to current time and any enabled timers. The assumption is
		// that if timers are set, the inverter's Power state may have changed to on or off
		// automatically, which is not reflected in the saved state, and we don't want to
		// inadvertently change it while changing some other parameter. Therefore Power will
		// be set to what is expected according to the timers. This automatic behavior can
		// be overridden by explicitly setting Power.
		// Note that rc may contain updated timer configuration - if not, fallback to using
		// saved values in dbRc.
		timer_on := rc.TimerOn == rcconst.C_Timer_Enabled
		on_time := rc.TimerOnTime
		if on_time == rcconst.C_Time_Unset {
			on_time = dbRc.TimerOnTime
		}
		timer_off := rc.TimerOff == rcconst.C_Timer_Enabled
		off_time := rc.TimerOffTime
		if off_time == rcconst.C_Time_Unset {
			off_time = dbRc.TimerOffTime
		}
		if timer_on && timer_off && on_time != rcconst.C_Time_Unset && off_time != rcconst.C_Time_Unset {
			if on_time <= off_time {
				rc.Power = rcconst.C_Power_On
				if clock < on_time || clock >= off_time {
					rc.Power = rcconst.C_Power_Off
				}
			} else {
				rc.Power = rcconst.C_Power_Off
				if clock < off_time || clock >= on_time {
					rc.Power = rcconst.C_Power_On
				}
			}
		} else if timer_on && on_time != rcconst.C_Time_Unset {
			rc.Power = rcconst.C_Power_Off
			if clock >= on_time {
				rc.Power = rcconst.C_Power_On
			}
		} else if timer_off && off_time != rcconst.C_Time_Unset {
			rc.Power = rcconst.C_Power_On
			if clock >= off_time {
				rc.Power = rcconst.C_Power_Off
			}
		}
	}
}

func SetMode(mode string, rc *codec.RcConfig) {
	switch mode {
	case "auto":
		rc.Mode = rcconst.C_Mode_Auto
	case "dry":
		rc.Mode = rcconst.C_Mode_Dry
	case "cool":
		rc.Mode = rcconst.C_Mode_Cool
	case "heat":
		rc.Mode = rcconst.C_Mode_Heat
	default:
		return
	}
	temp, fan, err := db.GetModeSettings(rc.Mode)
	if err != nil {
		return
	}
	if rc.Powerful == rcconst.C_Powerful_Disabled && rc.Quiet == rcconst.C_Quiet_Disabled {
		rc.FanSpeed = fan
	}
	rc.Temperature = temp
}

func SetPowerful(setting string, rc *codec.RcConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		rc.Powerful = rcconst.C_Powerful_Enabled
	case "off", "no", "disable", "disabled":
		rc.Powerful = rcconst.C_Powerful_Disabled
	default:
		return
	}
	if rc.Powerful == rcconst.C_Powerful_Enabled {
		rc.FanSpeed = rcconst.C_FanSpeed_Auto
		rc.Quiet = rcconst.C_Quiet_Disabled
	} else {
		_, fan, err := db.GetModeSettings(rc.Mode)
		if err != nil {
			return
		}
		rc.FanSpeed = fan
	}
}

func SetQuiet(setting string, rc *codec.RcConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		rc.Quiet = rcconst.C_Quiet_Enabled
	case "off", "no", "disable", "disabled":
		rc.Quiet = rcconst.C_Quiet_Disabled
	default:
		return
	}
	if rc.Quiet == rcconst.C_Quiet_Enabled {
		rc.FanSpeed = rcconst.C_FanSpeed_Lowest
		rc.Powerful = rcconst.C_Powerful_Disabled
	} else {
		_, fan, err := db.GetModeSettings(rc.Mode)
		if err != nil {
			return
		}
		rc.FanSpeed = fan
	}
}

func SetTemperature(temp string, rc *codec.RcConfig) {
	if t, err := strconv.Atoi(temp); err != nil {
		if rcconst.C_Temp_Min <= t && t <= rcconst.C_Temp_Max {
			rc.Temperature = uint(t)
		}
	}
}

func SetFanSpeed(fan string, rc *codec.RcConfig) {
	if rc.Powerful == rcconst.C_Powerful_Enabled || rc.Quiet == rcconst.C_Quiet_Enabled {
		return
	}
	switch fan {
	case "auto":
		rc.FanSpeed = rcconst.C_FanSpeed_Auto
	case "lowest", "slowest":
		rc.FanSpeed = rcconst.C_FanSpeed_Lowest
	case "low", "slow":
		rc.FanSpeed = rcconst.C_FanSpeed_Low
	case "middle", "center":
		rc.FanSpeed = rcconst.C_FanSpeed_Middle
	case "high", "fast":
		rc.FanSpeed = rcconst.C_FanSpeed_High
	case "highest", "fastest":
		rc.FanSpeed = rcconst.C_FanSpeed_Highest
	default:
		return
	}
}

func SetVentVerticalPosition(vert string, rc *codec.RcConfig) {
	switch vert {
	case "auto":
		rc.VentVertical = rcconst.C_VentVertical_Auto
	case "lowest", "bottom":
		rc.VentVertical = rcconst.C_VentVertical_Low
	case "low":
		rc.VentVertical = rcconst.C_VentVertical_Lowest
	case "middle", "center":
		rc.VentVertical = rcconst.C_VentVertical_Middle
	case "high":
		rc.VentVertical = rcconst.C_VentVertical_High
	case "highest", "top":
		rc.VentVertical = rcconst.C_VentVertical_Highest
	default:
		return
	}
}

func SetVentHorizontalPosition(horiz string, rc *codec.RcConfig) {
	switch horiz {
	case "auto":
		rc.VentHorizontal = rcconst.C_VentHorizontal_Auto
	case "farleft", "leftmost":
		rc.VentHorizontal = rcconst.C_VentHorizontal_FarLeft
	case "left":
		rc.VentHorizontal = rcconst.C_VentHorizontal_Left
	case "middle", "center":
		rc.VentHorizontal = rcconst.C_VentHorizontal_Middle
	case "right":
		rc.VentHorizontal = rcconst.C_VentHorizontal_Right
	case "farright", "rightmost":
		rc.VentHorizontal = rcconst.C_VentHorizontal_FarRight
	default:
		return
	}
}

// Timers

func parseTime(time string) (hour, minute int, err error) {
	v := strings.Split(time, ":")
	if len(v) != 2 {
		return hour, minute, fmt.Errorf("not a time")
	}
	hour, err = strconv.Atoi(v[0])
	if err != nil {
		return hour, minute, err
	}
	minute, err = strconv.Atoi(v[1])
	if err != nil {
		return hour, minute, err
	}
	if hour < 0 || hour > 23 {
		return hour, minute, fmt.Errorf("bad hour")
	}
	if minute < 0 || minute > 59 {
		return hour, minute, fmt.Errorf("bad minute")
	}
	return hour, minute, nil
}

func setTimes(rc, dbRc *codec.RcConfig) {
	// copy saved times if unset
	if rc.TimerOnTime == rcconst.C_Time_Unset {
		rc.TimerOnTime = dbRc.TimerOnTime
	}
	if rc.TimerOffTime == rcconst.C_Time_Unset {
		rc.TimerOffTime = dbRc.TimerOffTime
	}
	// set the clock field to the current time
	rc.SetClock()
}

func SetTimerOn(setting string, rc, dbRc *codec.RcConfig) {
	switch setting {
	case "on":
		rc.TimerOn = rcconst.C_Timer_Enabled
		setTimes(rc, dbRc)
	case "off":
		rc.TimerOn = rcconst.C_Timer_Disabled
		setTimes(rc, dbRc)
	default:
		return
	}
}

func SetTimerOff(setting string, rc, dbRc *codec.RcConfig) {
	switch setting {
	case "on":
		rc.TimerOff = rcconst.C_Timer_Enabled
		setTimes(rc, dbRc)
	case "off":
		rc.TimerOff = rcconst.C_Timer_Disabled
		setTimes(rc, dbRc)
	default:
		return
	}
}

func SetTimerOnTime(time string, rc, dbRc *codec.RcConfig) {
	if time == "" {
		return
	}
	hour, minute, err := parseTime(time)
	if err != nil {
		return
	}
	rc.TimerOnTime = codec.NewTime(uint(hour), uint(minute))
	setTimes(rc, dbRc)
}

func SetTimerOffTime(time string, rc, dbRc *codec.RcConfig) {
	if time == "" {
		return
	}
	hour, minute, err := parseTime(time)
	if err != nil {
		return
	}
	rc.TimerOffTime = codec.NewTime(uint(hour), uint(minute))
	setTimes(rc, dbRc)
}

func ComposeSendConfig(settings *rcconst.Settings, dbRc *codec.RcConfig) *codec.RcConfig {
	sendRc := dbRc.CopyForSending()
	SetMode(settings.Mode, sendRc)
	SetPowerful(settings.Powerful, sendRc)
	SetQuiet(settings.Quiet, sendRc)
	SetTemperature(settings.Temperature, sendRc)
	SetFanSpeed(settings.FanSpeed, sendRc)
	SetVentVerticalPosition(settings.VentVertical, sendRc)
	SetVentHorizontalPosition(settings.VentHorizontal, sendRc)

	// if timers are changed in any way, time fields are initialized
	SetTimerOn(settings.TimerOn, sendRc, dbRc)
	SetTimerOnTime(settings.TimerOnTime, sendRc, dbRc)
	SetTimerOff(settings.TimerOff, sendRc, dbRc)
	SetTimerOffTime(settings.TimerOffTime, sendRc, dbRc)

	// set power last, adjusting for any (updated) timers
	SetPower(settings.Power, sendRc, dbRc)

	return sendRc
}
