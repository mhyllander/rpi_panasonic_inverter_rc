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
	Active string // "true" or "false"
}

// Define cronjobs, their schedules, and which job set each cronjob belongs to.
type CronJob struct {
	gorm.Model
	JobSet   string // the job set that the cronjob belongs to
	Schedule string // schedule in crontab format
	Settings []byte // JSON representation of Settings struct
}

// Define a number of settings. Empty fields are unset. This struct is used
// to serialize settings to and from JSON, when they are stored in the db, and
// when they are sent to or received from web clients.
type Settings struct {
	Power          string `json:"power,omitempty"`
	Mode           string `json:"mode,omitempty"`
	Powerful       string `json:"powerful,omitempty"`
	Quiet          string `json:"quiet,omitempty"`
	Temperature    string `json:"temp,omitempty"`
	FanSpeed       string `json:"fan,omitempty"`
	VentVertical   string `json:"vert,omitempty"`
	VentHorizontal string `json:"horiz,omitempty"`
	TimerOn        string `json:"ton,omitempty"`
	TimerOnTime    string `json:"tont,omitempty"`
	TimerOff       string `json:"toff,omitempty"`
	TimerOffTime   string `json:"tofft,omitempty"`
}
