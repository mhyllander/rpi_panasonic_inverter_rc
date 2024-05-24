package codecbase

import "strconv"

// The inverter settings. Empty fields are unset. This struct is used to serialize settings to and from JSON, when they
// are stored as part of jobs in the database, and when they are sent to or received from web clients.
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

// The per-mode settings. Empty fields are unset. This struct is used to serialize settings to and from JSON when they
// are sent to or received from web clients.
type ModeSettings struct {
	Temperature string `json:"temp"`
	FanSpeed    string `json:"fan"`
}

type ModeSettingsMap map[string]ModeSettings

// Settings and all ModeSettings. This struct is used to pass all settings as JSON to a web client.
type AllSettings struct {
	Settings     Settings        `json:"settings"`
	ModeSettings ModeSettingsMap `json:"modeSettings"`
}

func Power2String(power uint) string {
	switch power {
	case C_Power_On:
		return "on"
	case C_Power_Off:
		return "off"
	}
	return ""
}

func Mode2String(mode uint) string {
	switch mode {
	case C_Mode_Auto:
		return "auto"
	case C_Mode_Heat:
		return "heat"
	case C_Mode_Cool:
		return "cool"
	case C_Mode_Dry:
		return "dry"
	}
	return ""
}

func Powerful2String(powerful uint) string {
	switch powerful {
	case C_Powerful_Enabled:
		return "on"
	case C_Powerful_Disabled:
		return "off"
	}
	return ""
}

func Quiet2String(quiet uint) string {
	switch quiet {
	case C_Quiet_Enabled:
		return "on"
	case C_Quiet_Disabled:
		return "off"
	}
	return ""
}

func Temperatur2String(temp uint) string {
	return strconv.Itoa(int(temp))
}

func FanSpeed2String(fan uint) string {
	switch fan {
	case C_FanSpeed_Auto:
		return "auto"
	case C_FanSpeed_Lowest:
		return "lowest"
	case C_FanSpeed_Low:
		return "low"
	case C_FanSpeed_Middle:
		return "middle"
	case C_FanSpeed_High:
		return "high"
	case C_FanSpeed_Highest:
		return "highest"
	}
	return ""
}

func VentVertical2String(vert uint) string {
	switch vert {
	case C_VentVertical_Auto:
		return "auto"
	case C_VentVertical_Lowest:
		return "lowest"
	case C_VentVertical_Low:
		return "low"
	case C_VentVertical_Middle:
		return "middle"
	case C_VentVertical_High:
		return "high"
	case C_VentVertical_Highest:
		return "highest"
	}
	return ""
}

func VentHorizontal2String(horiz uint) string {
	switch horiz {
	case C_VentHorizontal_Auto:
		return "auto"
	case C_VentHorizontal_FarLeft:
		return "farleft"
	case C_VentHorizontal_Left:
		return "left"
	case C_VentHorizontal_Middle:
		return "middle"
	case C_VentHorizontal_Right:
		return "right"
	case C_VentHorizontal_FarRight:
		return "farright"
	}
	return ""
}

func TimerToString(t uint) string {
	switch t {
	case C_Timer_Enabled:
		return "on"
	case C_Timer_Disabled:
		return "off"
	}
	return ""
}
