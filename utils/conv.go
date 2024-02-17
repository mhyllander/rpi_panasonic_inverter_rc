package utils

import (
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/common"
)

func CopyToSettings(rc *codec.RcConfig, settings *common.Settings) {
	settings.Power = common.Power2String(rc.Power)
	settings.Mode = common.Mode2String(rc.Mode)
	settings.Powerful = common.Powerful2String(rc.Powerful)
	settings.Quiet = common.Quiet2String(rc.Quiet)
	settings.Temperature = common.Temperatur2String(rc.Temperature)
	settings.FanSpeed = common.FanSpeed2String(rc.FanSpeed)
	settings.VentVertical = common.VentVertical2String(rc.VentVertical)
	settings.VentHorizontal = common.VentHorizontal2String(rc.VentHorizontal)
	settings.TimerOn = common.TimerToString(rc.TimerOn)
	settings.TimerOnTime = rc.TimerOnTime.ToString()
	settings.TimerOff = common.TimerToString(rc.TimerOff)
	settings.TimerOffTime = rc.TimerOffTime.ToString()
}

func CopyToModeSettings(temp, fan uint, modeSettings *common.ModeSettings) {
	modeSettings.FanSpeed = common.FanSpeed2String(fan)
	modeSettings.Temperature = common.Temperatur2String(temp)
}
