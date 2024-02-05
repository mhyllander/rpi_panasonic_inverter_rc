package rcconst

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
