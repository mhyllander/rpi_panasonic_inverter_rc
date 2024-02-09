package db

import (
	"gorm.io/gorm"
)

type DbIrConfig struct {
	gorm.Model
	Power          uint
	Mode           uint
	Powerful       uint
	Quiet          uint
	Temperature    uint
	FanSpeed       uint
	VentVertical   uint
	VentHorizontal uint
	TimerOn        uint
	TimerOff       uint
	TimerOnTime    uint
	TimerOffTime   uint
}

type ModeSetting struct {
	gorm.Model
	Mode        uint
	Temperature uint
	FanSpeed    uint
}

// Define sets of cronjobs, e.g. Normal, Vacation, Home, Away, etc.
// This allows toggling which cronjobs are active.
type JobSet struct {
	gorm.Model
	Name   string // name of the job set
	Active bool   // true or false
}

// Define cronjobs, their schedules, and which job set each cronjob belongs to.
type CronJob struct {
	gorm.Model
	JobSet   string // the job set that the cronjob belongs to
	Schedule string // schedule in crontab format
	Settings []byte // JSON representation of Settings struct
}
