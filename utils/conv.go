package utils

import (
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
)

func CopyToSettings(rc *codec.RcConfig, settings *codecbase.Settings) {
	settings.Power = codecbase.Power2String(rc.Power)
	settings.Mode = codecbase.Mode2String(rc.Mode)
	settings.Powerful = codecbase.Powerful2String(rc.Powerful)
	settings.Quiet = codecbase.Quiet2String(rc.Quiet)
	settings.Temperature = codecbase.Temperatur2String(rc.Temperature)
	settings.FanSpeed = codecbase.FanSpeed2String(rc.FanSpeed)
	settings.VentVertical = codecbase.VentVertical2String(rc.VentVertical)
	settings.VentHorizontal = codecbase.VentHorizontal2String(rc.VentHorizontal)
	settings.TimerOn = codecbase.TimerToString(rc.TimerOn)
	settings.TimerOnTime = rc.TimerOnTime.ToString()
	settings.TimerOff = codecbase.TimerToString(rc.TimerOff)
	settings.TimerOffTime = rc.TimerOffTime.ToString()
}

func CopyToModeSettings(temp, fan uint, modeSettings *codecbase.ModeSettings) {
	modeSettings.FanSpeed = codecbase.FanSpeed2String(fan)
	modeSettings.Temperature = codecbase.Temperatur2String(temp)
}
