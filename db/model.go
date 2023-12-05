package db

import (
	"gorm.io/gorm"
)

type DbIrConfig struct {
	gorm.Model
	Power           uint
	Mode            uint
	Powerful        uint
	Quiet           uint
	Temperature     uint
	FanSpeed        uint
	VentVertical    uint
	VentHorizontal  uint
	TimerOnEnabled  uint
	TimerOffEnabled uint
	TimerOn         uint
	TimerOff        uint
}

type ModeSetting struct {
	gorm.Model
	Mode        uint
	Temperature uint
	FanSpeed    uint
}
