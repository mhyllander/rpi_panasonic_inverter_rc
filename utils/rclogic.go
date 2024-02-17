package utils

import (
	"fmt"
	"log/slog"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/common"
	"rpi_panasonic_inverter_rc/db"
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
		rc.Power = common.C_Power_On
	case "off", "no", "disable", "disabled":
		rc.Power = common.C_Power_Off
	default:
		// Adjust Power according to current time and any enabled timers. The assumption is
		// that if timers are set, the inverter's Power state may have changed to on or off
		// automatically, which is not reflected in the saved state, and we don't want to
		// inadvertently change it while changing some other parameter. Therefore Power will
		// be set to what is expected according to the timers. This automatic behavior can
		// be overridden by explicitly setting Power.
		// Note that rc may contain updated timer configuration - if not, fallback to using
		// saved values in dbRc.
		timer_on := rc.TimerOn == common.C_Timer_Enabled
		on_time := rc.TimerOnTime
		if on_time == common.C_Time_Unset {
			on_time = dbRc.TimerOnTime
		}
		timer_off := rc.TimerOff == common.C_Timer_Enabled
		off_time := rc.TimerOffTime
		if off_time == common.C_Time_Unset {
			off_time = dbRc.TimerOffTime
		}
		if timer_on && timer_off && on_time != common.C_Time_Unset && off_time != common.C_Time_Unset {
			if on_time <= off_time {
				rc.Power = common.C_Power_On
				if clock < on_time || clock >= off_time {
					rc.Power = common.C_Power_Off
				}
			} else {
				rc.Power = common.C_Power_Off
				if clock < off_time || clock >= on_time {
					rc.Power = common.C_Power_On
				}
			}
		} else if timer_on && on_time != common.C_Time_Unset {
			rc.Power = common.C_Power_Off
			if clock >= on_time {
				rc.Power = common.C_Power_On
			}
		} else if timer_off && off_time != common.C_Time_Unset {
			rc.Power = common.C_Power_On
			if clock >= off_time {
				rc.Power = common.C_Power_Off
			}
		}
	}
}

func SetMode(mode string, rc *codec.RcConfig) {
	switch mode {
	case "auto":
		rc.Mode = common.C_Mode_Auto
	case "dry":
		rc.Mode = common.C_Mode_Dry
	case "cool":
		rc.Mode = common.C_Mode_Cool
	case "heat":
		rc.Mode = common.C_Mode_Heat
	default:
		return
	}
	temp, fan, err := db.GetModeSettings(rc.Mode)
	if err != nil {
		return
	}
	if rc.Powerful == common.C_Powerful_Disabled && rc.Quiet == common.C_Quiet_Disabled {
		rc.FanSpeed = fan
	}
	rc.Temperature = temp
}

func SetPowerful(setting string, rc *codec.RcConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		rc.Powerful = common.C_Powerful_Enabled
	case "off", "no", "disable", "disabled":
		rc.Powerful = common.C_Powerful_Disabled
	default:
		return
	}
	if rc.Powerful == common.C_Powerful_Enabled {
		rc.FanSpeed = common.C_FanSpeed_Auto
		rc.Quiet = common.C_Quiet_Disabled
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
		rc.Quiet = common.C_Quiet_Enabled
	case "off", "no", "disable", "disabled":
		rc.Quiet = common.C_Quiet_Disabled
	default:
		return
	}
	if rc.Quiet == common.C_Quiet_Enabled {
		rc.FanSpeed = common.C_FanSpeed_Lowest
		rc.Powerful = common.C_Powerful_Disabled
	} else {
		_, fan, err := db.GetModeSettings(rc.Mode)
		if err != nil {
			return
		}
		rc.FanSpeed = fan
	}
}

func SetTemperature(temp string, rc *codec.RcConfig) {
	if temp == "" {
		return
	}
	if t, err := strconv.Atoi(temp); err == nil {
		if common.C_Temp_Min <= t && t <= common.C_Temp_Max {
			rc.Temperature = uint(t)
		}
	} else {
		slog.Warn("cannot convert temperature", "temp", temp, "err", err)
	}
}

func SetFanSpeed(fan string, rc *codec.RcConfig) {
	if rc.Powerful == common.C_Powerful_Enabled || rc.Quiet == common.C_Quiet_Enabled {
		return
	}
	switch fan {
	case "auto":
		rc.FanSpeed = common.C_FanSpeed_Auto
	case "lowest", "slowest":
		rc.FanSpeed = common.C_FanSpeed_Lowest
	case "low", "slow":
		rc.FanSpeed = common.C_FanSpeed_Low
	case "middle", "center":
		rc.FanSpeed = common.C_FanSpeed_Middle
	case "high", "fast":
		rc.FanSpeed = common.C_FanSpeed_High
	case "highest", "fastest":
		rc.FanSpeed = common.C_FanSpeed_Highest
	default:
		return
	}
}

func SetVentVerticalPosition(vert string, rc *codec.RcConfig) {
	switch vert {
	case "auto":
		rc.VentVertical = common.C_VentVertical_Auto
	case "lowest", "bottom":
		rc.VentVertical = common.C_VentVertical_Lowest
	case "low":
		rc.VentVertical = common.C_VentVertical_Low
	case "middle", "center":
		rc.VentVertical = common.C_VentVertical_Middle
	case "high":
		rc.VentVertical = common.C_VentVertical_High
	case "highest", "top":
		rc.VentVertical = common.C_VentVertical_Highest
	default:
		return
	}
}

func SetVentHorizontalPosition(horiz string, rc *codec.RcConfig) {
	switch horiz {
	case "auto":
		rc.VentHorizontal = common.C_VentHorizontal_Auto
	case "farleft", "leftmost":
		rc.VentHorizontal = common.C_VentHorizontal_FarLeft
	case "left":
		rc.VentHorizontal = common.C_VentHorizontal_Left
	case "middle", "center":
		rc.VentHorizontal = common.C_VentHorizontal_Middle
	case "right":
		rc.VentHorizontal = common.C_VentHorizontal_Right
	case "farright", "rightmost":
		rc.VentHorizontal = common.C_VentHorizontal_FarRight
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
	if rc.TimerOnTime == common.C_Time_Unset {
		rc.TimerOnTime = dbRc.TimerOnTime
	}
	if rc.TimerOffTime == common.C_Time_Unset {
		rc.TimerOffTime = dbRc.TimerOffTime
	}
	// set the clock field to the current time
	rc.SetClock()
}

func SetTimerOn(setting string, rc, dbRc *codec.RcConfig) {
	switch setting {
	case "on":
		rc.TimerOn = common.C_Timer_Enabled
		setTimes(rc, dbRc)
	case "off":
		rc.TimerOn = common.C_Timer_Disabled
		setTimes(rc, dbRc)
	default:
		return
	}
}

func SetTimerOff(setting string, rc, dbRc *codec.RcConfig) {
	switch setting {
	case "on":
		rc.TimerOff = common.C_Timer_Enabled
		setTimes(rc, dbRc)
	case "off":
		rc.TimerOff = common.C_Timer_Disabled
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

func ComposeSendConfig(settings *common.Settings, dbRc *codec.RcConfig) *codec.RcConfig {
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
