package utils

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
	"rpi_panasonic_inverter_rc/db"
)

func SetPower(setting string, rc *codec.RcConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		rc.Power = codecbase.C_Power_On
	case "off", "no", "disable", "disabled":
		rc.Power = codecbase.C_Power_Off
	default:
		return
	}
}

func SetMode(mode string, rc *codec.RcConfig) {
	switch mode {
	case "auto":
		rc.Mode = codecbase.C_Mode_Auto
	case "dry":
		rc.Mode = codecbase.C_Mode_Dry
	case "cool":
		rc.Mode = codecbase.C_Mode_Cool
	case "heat":
		rc.Mode = codecbase.C_Mode_Heat
	default:
		return
	}
	temp, fan, err := db.GetModeSettings(rc.Mode)
	if err != nil {
		return
	}
	if rc.Powerful == codecbase.C_Powerful_Disabled && rc.Quiet == codecbase.C_Quiet_Disabled {
		rc.FanSpeed = fan
	}
	rc.Temperature = temp
}

func SetPowerful(setting string, rc *codec.RcConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		rc.Powerful = codecbase.C_Powerful_Enabled
	case "off", "no", "disable", "disabled":
		rc.Powerful = codecbase.C_Powerful_Disabled
	default:
		return
	}
	if rc.Powerful == codecbase.C_Powerful_Enabled {
		rc.FanSpeed = codecbase.C_FanSpeed_Auto
		rc.Quiet = codecbase.C_Quiet_Disabled
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
		rc.Quiet = codecbase.C_Quiet_Enabled
	case "off", "no", "disable", "disabled":
		rc.Quiet = codecbase.C_Quiet_Disabled
	default:
		return
	}
	if rc.Quiet == codecbase.C_Quiet_Enabled {
		rc.FanSpeed = codecbase.C_FanSpeed_Lowest
		rc.Powerful = codecbase.C_Powerful_Disabled
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
		if codecbase.C_Temp_Min <= t && t <= codecbase.C_Temp_Max {
			rc.Temperature = uint(t)
		}
	} else {
		slog.Warn("cannot convert temperature", "temp", temp, "err", err)
	}
}

func SetFanSpeed(fan string, rc *codec.RcConfig) {
	if rc.Powerful == codecbase.C_Powerful_Enabled || rc.Quiet == codecbase.C_Quiet_Enabled {
		return
	}
	switch fan {
	case "auto":
		rc.FanSpeed = codecbase.C_FanSpeed_Auto
	case "lowest", "slowest":
		rc.FanSpeed = codecbase.C_FanSpeed_Lowest
	case "low", "slow":
		rc.FanSpeed = codecbase.C_FanSpeed_Low
	case "middle", "center":
		rc.FanSpeed = codecbase.C_FanSpeed_Middle
	case "high", "fast":
		rc.FanSpeed = codecbase.C_FanSpeed_High
	case "highest", "fastest":
		rc.FanSpeed = codecbase.C_FanSpeed_Highest
	default:
		return
	}
}

func SetVentVerticalPosition(vert string, rc *codec.RcConfig) {
	switch vert {
	case "auto":
		rc.VentVertical = codecbase.C_VentVertical_Auto
	case "lowest", "bottom":
		rc.VentVertical = codecbase.C_VentVertical_Lowest
	case "low":
		rc.VentVertical = codecbase.C_VentVertical_Low
	case "middle", "center":
		rc.VentVertical = codecbase.C_VentVertical_Middle
	case "high":
		rc.VentVertical = codecbase.C_VentVertical_High
	case "highest", "top":
		rc.VentVertical = codecbase.C_VentVertical_Highest
	default:
		return
	}
}

func SetVentHorizontalPosition(horiz string, rc *codec.RcConfig) {
	switch horiz {
	case "auto":
		rc.VentHorizontal = codecbase.C_VentHorizontal_Auto
	case "farleft", "leftmost":
		rc.VentHorizontal = codecbase.C_VentHorizontal_FarLeft
	case "left":
		rc.VentHorizontal = codecbase.C_VentHorizontal_Left
	case "middle", "center":
		rc.VentHorizontal = codecbase.C_VentHorizontal_Middle
	case "right":
		rc.VentHorizontal = codecbase.C_VentHorizontal_Right
	case "farright", "rightmost":
		rc.VentHorizontal = codecbase.C_VentHorizontal_FarRight
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
	if rc.TimerOnTime == codecbase.C_Time_Unset {
		rc.TimerOnTime = dbRc.TimerOnTime
	}
	if rc.TimerOffTime == codecbase.C_Time_Unset {
		rc.TimerOffTime = dbRc.TimerOffTime
	}
	// set the clock field to the current time
	rc.SetClock()
}

func SetTimerOn(setting string, rc, dbRc *codec.RcConfig) {
	switch setting {
	case "on":
		rc.TimerOn = codecbase.C_Timer_Enabled
		setTimes(rc, dbRc)
	case "off":
		rc.TimerOn = codecbase.C_Timer_Disabled
		setTimes(rc, dbRc)
	default:
		return
	}
}

func SetTimerOff(setting string, rc, dbRc *codec.RcConfig) {
	switch setting {
	case "on":
		rc.TimerOff = codecbase.C_Timer_Enabled
		setTimes(rc, dbRc)
	case "off":
		rc.TimerOff = codecbase.C_Timer_Disabled
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

func ComposeSendConfig(settings *codecbase.Settings, dbRc *codec.RcConfig) *codec.RcConfig {
	sendRc := dbRc.CopyForSending()
	SetPower(settings.Power, sendRc)
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

	return sendRc
}
