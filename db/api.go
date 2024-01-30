package db

import (
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/rcconst"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var myDb *gorm.DB

func GetDBPath() string {
	db := os.Getenv("PANINV_DB")
	if db == "" {
		db = "paninv.db"
	}
	return db
}

func Initialize(dbFile string) error {
	var err error
	myDb, err = gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	myDb.AutoMigrate(&DbIrConfig{})
	myDb.AutoMigrate(&ModeSetting{})

	// Create initial records
	var dbRc DbIrConfig

	result := myDb.First(&dbRc, 1)
	if result.Error != nil && result.RowsAffected == 0 {
		slog.Info("Initializing db")

		rc := codec.NewRcConfig()
		c := &DbIrConfig{
			Power:          rc.Power,
			Mode:           rc.Mode,
			Powerful:       rc.Powerful,
			Quiet:          rc.Quiet,
			Temperature:    rc.Temperature,
			FanSpeed:       rc.FanSpeed,
			VentVertical:   rc.VentVertical,
			VentHorizontal: rc.VentHorizontal,
			TimerOn:        rc.TimerOn,
			TimerOff:       rc.TimerOff,
			TimerOnTime:    uint(rc.TimerOnTime),
			TimerOffTime:   uint(rc.TimerOffTime),
		}
		result := myDb.Create(c)
		if result.Error != nil {
			return err
		}

		settings := []*ModeSetting{
			{Mode: rcconst.C_Mode_Auto, Temperature: 21, FanSpeed: rcconst.C_FanSpeed_Auto},
			{Mode: rcconst.C_Mode_Dry, Temperature: 20, FanSpeed: rcconst.C_FanSpeed_Auto},
			{Mode: rcconst.C_Mode_Heat, Temperature: 24, FanSpeed: rcconst.C_FanSpeed_Auto},
			{Mode: rcconst.C_Mode_Cool, Temperature: 18, FanSpeed: rcconst.C_FanSpeed_Auto},
		}
		result = myDb.Create(settings)
		if result.Error != nil {
			return err
		}
	}

	return nil
}

func Close() {

}

func CurrentConfig() (*codec.RcConfig, error) {
	var dbRc DbIrConfig
	result := myDb.First(&dbRc, 1)
	if result.Error != nil {
		return nil, result.Error
	}
	return &codec.RcConfig{
		Power:          dbRc.Power,
		Mode:           dbRc.Mode,
		Powerful:       dbRc.Powerful,
		Quiet:          dbRc.Quiet,
		Temperature:    dbRc.Temperature,
		FanSpeed:       dbRc.FanSpeed,
		VentVertical:   dbRc.VentVertical,
		VentHorizontal: dbRc.VentHorizontal,
		TimerOn:        dbRc.TimerOn,
		TimerOff:       dbRc.TimerOff,
		TimerOnTime:    codec.Time(dbRc.TimerOnTime),
		TimerOffTime:   codec.Time(dbRc.TimerOffTime),
		Clock:          rcconst.C_Time_Unset,
	}, nil
}

// Return the current config with initialized clock, useful for
// immediately sending the current config to the inverter after e.g.
// a power outage. It could be used to send the current configuration
// after the RPi has booted up.
func CurrentConfigForSending() (*codec.RcConfig, error) {
	dbRc, err := CurrentConfig()
	if err != nil {
		return dbRc, err
	}
	dbRc.SetClock()
	return dbRc, nil
}

func SaveConfig(rc, dbRc *codec.RcConfig) error {
	// update current configuration, but timer on and off should only be updated if set
	// mode settings should be updated, but fan speed should be ignored if Powerful or Quiet is set

	var updates, settings map[string]interface{}
	updates = make(map[string]interface{})
	settings = make(map[string]interface{})

	if rc.Power != dbRc.Power {
		updates["Power"] = rc.Power
	}
	if rc.Mode != dbRc.Mode {
		updates["Mode"] = rc.Mode
	}
	if rc.Powerful != dbRc.Powerful {
		updates["Powerful"] = rc.Powerful
	}
	if rc.Quiet != dbRc.Quiet {
		updates["Quiet"] = rc.Quiet
	}
	if rc.Temperature != dbRc.Temperature {
		updates["Temperature"] = rc.Temperature
	}
	if rc.FanSpeed != dbRc.FanSpeed {
		updates["FanSpeed"] = rc.FanSpeed
	}
	if rc.VentVertical != dbRc.VentVertical {
		updates["VentVertical"] = rc.VentVertical
	}
	if rc.VentHorizontal != dbRc.VentHorizontal {
		updates["VentHorizontal"] = rc.VentHorizontal
	}
	if rc.TimerOn != dbRc.TimerOn {
		updates["TimerOn"] = rc.TimerOn
	}
	if rc.TimerOff != dbRc.TimerOff {
		updates["TimerOff"] = rc.TimerOff
	}
	if rc.TimerOnTime != dbRc.TimerOnTime && rc.TimerOnTime != rcconst.C_Time_Unset {
		updates["TimerOnTime"] = rc.TimerOnTime
	}
	if rc.TimerOffTime != dbRc.TimerOffTime && rc.TimerOffTime != rcconst.C_Time_Unset {
		updates["TimerOffTime"] = rc.TimerOffTime
	}

	var nc DbIrConfig
	myDb.First(&nc, 1)
	result := myDb.Model(&nc).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	settings["Temperature"] = rc.Temperature
	if rc.Powerful == rcconst.C_Powerful_Disabled && rc.Quiet == rcconst.C_Quiet_Disabled {
		settings["FanSpeed"] = rc.FanSpeed
	}
	result = myDb.Model(ModeSetting{}).Where("mode = ?", rc.Mode).Updates(settings)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func GetModeSettings(mode uint) (temp, fan uint, err error) {
	var ms ModeSetting
	result := myDb.First(&ms, "mode = ?", mode)
	if result.Error != nil {
		return 0, 0, result.Error
	}
	return ms.Temperature, ms.FanSpeed, nil
}
