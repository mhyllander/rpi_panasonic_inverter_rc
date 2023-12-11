package rclogic

import (
	"fmt"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"strconv"
	"strings"
	"time"
)

func SetPower(setting string, ic *codec.IrConfig) {
	switch setting {
	case "on":
		ic.Power = codec.C_Power_On
	case "off":
		ic.Power = codec.C_Power_Off
	default:
		return
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
	case "on":
		ic.Powerful = codec.C_Powerful_Enabled
	case "off":
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
	case "on":
		ic.Quiet = codec.C_Quiet_Enabled
	case "off":
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
	case "middle":
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
	case "middle":
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
	case "farleft":
		ic.VentHorizontal = codec.C_VentHorizontal_FarLeft
	case "left":
		ic.VentHorizontal = codec.C_VentHorizontal_Left
	case "middle":
		ic.VentHorizontal = codec.C_VentHorizontal_Middle
	case "right":
		ic.VentHorizontal = codec.C_VentHorizontal_Right
	case "farright":
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
	if ic.TimerOn == codec.C_Time_Unset {
		ic.TimerOn = dbIc.TimerOn
	}
	if ic.TimerOff == codec.C_Time_Unset {
		ic.TimerOff = dbIc.TimerOff
	}
	// the the clock field to the current time
	now := time.Now()
	ic.Clock = codec.NewTime(uint(now.Hour()), uint(now.Minute()))
}

func SetTimerOnEnabled(setting string, ic, dbIc *codec.IrConfig) {
	switch setting {
	case "on":
		ic.TimerOnEnabled = codec.C_Timer_Enabled
		setTimes(ic, dbIc)
	case "off":
		ic.TimerOnEnabled = codec.C_Timer_Disabled
		setTimes(ic, dbIc)
	default:
		return
	}
}

func SetTimerOffEnabled(setting string, ic, dbIc *codec.IrConfig) {
	switch setting {
	case "on":
		ic.TimerOffEnabled = codec.C_Timer_Enabled
		setTimes(ic, dbIc)
	case "off":
		ic.TimerOffEnabled = codec.C_Timer_Disabled
		setTimes(ic, dbIc)
	default:
		return
	}
}

func SetTimerOn(time string, ic, dbIc *codec.IrConfig) {
	if time == "" {
		return
	}
	hour, minute, err := parseTime(time)
	if err != nil {
		return
	}
	ic.TimerOn = codec.NewTime(uint(hour), uint(minute))
	setTimes(ic, dbIc)
}

func SetTimerOff(time string, ic, dbIc *codec.IrConfig) {
	if time == "" {
		return
	}
	hour, minute, err := parseTime(time)
	if err != nil {
		return
	}
	ic.TimerOff = codec.NewTime(uint(hour), uint(minute))
	setTimes(ic, dbIc)
}
