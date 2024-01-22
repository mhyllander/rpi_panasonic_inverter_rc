package utils

import (
	"fmt"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"strconv"
	"strings"
	"time"
)

func SetPower(setting string, ic, dbIc *codec.IrConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		ic.Power = codec.C_Power_On
	case "off", "no", "disable", "disabled":
		ic.Power = codec.C_Power_Off
	default:
		// Adjust Power according to current time and any enabled timers. The assumption is
		// that if timers are set, the inverter's Power state may have changed to on or off
		// automatically, which is not reflected in the saved state, and we don't want to
		// inadvertently change it while changing some other parameter.
		// Note that ic contains the timer times to send, which may be unset, therefore
		// we also need to check any saved values in dbIc.
		now := time.Now()
		clock := codec.NewTime(uint(now.Hour()), uint(now.Minute()))
		if ic.TimerOn == codec.C_Timer_Enabled && ic.TimerOnTime != codec.C_Time_Unset && clock >= ic.TimerOnTime ||
			dbIc.TimerOn == codec.C_Timer_Enabled && dbIc.TimerOnTime != codec.C_Time_Unset && clock >= dbIc.TimerOnTime {
			ic.Power = codec.C_Power_On
		}
		if ic.TimerOff == codec.C_Timer_Enabled && ic.TimerOffTime != codec.C_Time_Unset && clock >= ic.TimerOffTime ||
			dbIc.TimerOff == codec.C_Timer_Enabled && dbIc.TimerOffTime != codec.C_Time_Unset && clock >= dbIc.TimerOffTime {
			ic.Power = codec.C_Power_Off
		}
	}
}

func SetMode(mode string, ic *codec.IrConfig) {
	switch mode {
	case "auto":
		ic.Mode = codec.C_Mode_Auto
	case "dry":
		ic.Mode = codec.C_Mode_Dry
	case "cool":
		ic.Mode = codec.C_Mode_Cool
	case "heat":
		ic.Mode = codec.C_Mode_Heat
	default:
		return
	}
	temp, fan, err := db.GetModeSettings(ic.Mode)
	if err != nil {
		return
	}
	if ic.Powerful == codec.C_Powerful_Disabled && ic.Quiet == codec.C_Quiet_Disabled {
		ic.FanSpeed = fan
	}
	ic.Temperature = temp
}

func SetPowerful(setting string, ic *codec.IrConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		ic.Powerful = codec.C_Powerful_Enabled
	case "off", "no", "disable", "disabled":
		ic.Powerful = codec.C_Powerful_Disabled
	default:
		return
	}
	if ic.Powerful == codec.C_Powerful_Enabled {
		ic.FanSpeed = codec.C_FanSpeed_Auto
		ic.Quiet = codec.C_Quiet_Disabled
	} else {
		_, fan, err := db.GetModeSettings(ic.Mode)
		if err != nil {
			return
		}
		ic.FanSpeed = fan
	}
}

func SetQuiet(setting string, ic *codec.IrConfig) {
	switch setting {
	case "on", "yes", "enable", "enabled":
		ic.Quiet = codec.C_Quiet_Enabled
	case "off", "no", "disable", "disabled":
		ic.Quiet = codec.C_Quiet_Disabled
	default:
		return
	}
	if ic.Quiet == codec.C_Quiet_Enabled {
		ic.FanSpeed = codec.C_FanSpeed_Lowest
		ic.Powerful = codec.C_Powerful_Disabled
	} else {
		_, fan, err := db.GetModeSettings(ic.Mode)
		if err != nil {
			return
		}
		ic.FanSpeed = fan
	}
}

func SetTemperature(temp int, ic *codec.IrConfig) {
	if codec.C_Temp_Min <= temp && temp <= codec.C_Temp_Max {
		ic.Temperature = uint(temp)
	}
}

func SetFanSpeed(fan string, ic *codec.IrConfig) {
	if ic.Powerful == codec.C_Powerful_Enabled || ic.Quiet == codec.C_Quiet_Enabled {
		return
	}
	switch fan {
	case "auto":
		ic.FanSpeed = codec.C_FanSpeed_Auto
	case "lowest":
		ic.FanSpeed = codec.C_FanSpeed_Lowest
	case "low":
		ic.FanSpeed = codec.C_FanSpeed_Low
	case "middle", "center":
		ic.FanSpeed = codec.C_FanSpeed_Middle
	case "high":
		ic.FanSpeed = codec.C_FanSpeed_High
	case "highest":
		ic.FanSpeed = codec.C_FanSpeed_Highest
	default:
		return
	}
}

func SetVentVerticalPosition(vert string, ic *codec.IrConfig) {
	switch vert {
	case "auto":
		ic.VentVertical = codec.C_VentVertical_Auto
	case "lowest":
		ic.VentVertical = codec.C_VentVertical_Low
	case "low":
		ic.VentVertical = codec.C_VentVertical_Lowest
	case "middle", "center":
		ic.VentVertical = codec.C_VentVertical_Middle
	case "high":
		ic.VentVertical = codec.C_VentVertical_High
	case "highest":
		ic.VentVertical = codec.C_VentVertical_Highest
	default:
		return
	}
}

func SetVentHorizontalPosition(horiz string, ic *codec.IrConfig) {
	switch horiz {
	case "auto":
		ic.VentHorizontal = codec.C_VentHorizontal_Auto
	case "farleft", "leftmost":
		ic.VentHorizontal = codec.C_VentHorizontal_FarLeft
	case "left":
		ic.VentHorizontal = codec.C_VentHorizontal_Left
	case "middle", "center":
		ic.VentHorizontal = codec.C_VentHorizontal_Middle
	case "right":
		ic.VentHorizontal = codec.C_VentHorizontal_Right
	case "farright", "rightmost":
		ic.VentHorizontal = codec.C_VentHorizontal_FarRight
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

func setTimes(ic, dbIc *codec.IrConfig) {
	// copy saved times if unset
	if ic.TimerOnTime == codec.C_Time_Unset {
		ic.TimerOnTime = dbIc.TimerOnTime
	}
	if ic.TimerOffTime == codec.C_Time_Unset {
		ic.TimerOffTime = dbIc.TimerOffTime
	}
	// set the clock field to the current time
	now := time.Now()
	ic.Clock = codec.NewTime(uint(now.Hour()), uint(now.Minute()))
}

func SetTimerOn(setting string, ic, dbIc *codec.IrConfig) {
	switch setting {
	case "on":
		ic.TimerOn = codec.C_Timer_Enabled
		setTimes(ic, dbIc)
	case "off":
		ic.TimerOn = codec.C_Timer_Disabled
		setTimes(ic, dbIc)
	default:
		return
	}
}

func SetTimerOff(setting string, ic, dbIc *codec.IrConfig) {
	switch setting {
	case "on":
		ic.TimerOff = codec.C_Timer_Enabled
		setTimes(ic, dbIc)
	case "off":
		ic.TimerOff = codec.C_Timer_Disabled
		setTimes(ic, dbIc)
	default:
		return
	}
}

func SetTimerOnTime(time string, ic, dbIc *codec.IrConfig) {
	if time == "" {
		return
	}
	hour, minute, err := parseTime(time)
	if err != nil {
		return
	}
	ic.TimerOnTime = codec.NewTime(uint(hour), uint(minute))
	setTimes(ic, dbIc)
}

func SetTimerOffTime(time string, ic, dbIc *codec.IrConfig) {
	if time == "" {
		return
	}
	hour, minute, err := parseTime(time)
	if err != nil {
		return
	}
	ic.TimerOffTime = codec.NewTime(uint(hour), uint(minute))
	setTimes(ic, dbIc)
}
