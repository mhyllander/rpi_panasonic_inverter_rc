package db

import (
	"fmt"
	"rpi_panasonic_inverter_rc/codec"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var myDb *gorm.DB

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
	var dbc DbIrConfig

	result := myDb.First(&dbc, 1)
	if result.Error != nil && result.RowsAffected == 0 {
		fmt.Println("Initializing db")

		ic := codec.NewIrConfig(nil)
		c := &DbIrConfig{
			Power:           ic.Power,
			Mode:            ic.Mode,
			Powerful:        ic.Powerful,
			Quiet:           ic.Quiet,
			Temperature:     ic.Temperature,
			FanSpeed:        ic.FanSpeed,
			VentVertical:    ic.VentVertical,
			VentHorizontal:  ic.VentHorizontal,
			TimerOnEnabled:  ic.TimerOnEnabled,
			TimerOffEnabled: ic.TimerOffEnabled,
			TimerOn:         uint(ic.TimerOn),
			TimerOff:        uint(ic.TimerOff),
		}
		result := myDb.Create(c)
		if result.Error != nil {
			return err
		}

		settings := []*ModeSetting{
			{Mode: codec.C_Mode_Auto, Temperature: 21, FanSpeed: codec.C_FanSpeed_Auto},
			{Mode: codec.C_Mode_Dry, Temperature: 20, FanSpeed: codec.C_FanSpeed_Auto},
			{Mode: codec.C_Mode_Heat, Temperature: 24, FanSpeed: codec.C_FanSpeed_Auto},
			{Mode: codec.C_Mode_Cool, Temperature: 18, FanSpeed: codec.C_FanSpeed_Auto},
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

func CurrentConfig() (*codec.IrConfig, error) {
	var dbc DbIrConfig
	result := myDb.First(&dbc, 1)
	if result.Error != nil {
		return nil, result.Error
	}
	return &codec.IrConfig{
		Power:           dbc.Power,
		Mode:            dbc.Mode,
		Powerful:        dbc.Powerful,
		Quiet:           dbc.Quiet,
		Temperature:     dbc.Temperature,
		FanSpeed:        dbc.FanSpeed,
		VentVertical:    dbc.VentVertical,
		VentHorizontal:  dbc.VentHorizontal,
		TimerOnEnabled:  dbc.TimerOnEnabled,
		TimerOffEnabled: dbc.TimerOffEnabled,
		TimerOn:         codec.Time(dbc.TimerOn),
		TimerOff:        codec.Time(dbc.TimerOff),
		Clock:           codec.C_Time_Unset,
	}, nil
}

func SaveConfig(ic, dbc *codec.IrConfig) error {
	// update current configuration, but timer on and off should only be updated if set
	// mode settings should be updated, but fan speed should be ignored if Powerful or Quiet is set

	var updates, settings map[string]interface{}
	updates = make(map[string]interface{})
	settings = make(map[string]interface{})

	if ic.Power != dbc.Power {
		updates["Power"] = ic.Power
	}
	if ic.Mode != dbc.Mode {
		updates["Mode"] = ic.Mode
	}
	if ic.Powerful != dbc.Powerful {
		updates["Powerful"] = ic.Powerful
	}
	if ic.Quiet != dbc.Quiet {
		updates["Quiet"] = ic.Quiet
	}
	if ic.Temperature != dbc.Temperature {
		updates["Temperature"] = ic.Temperature
	}
	if ic.FanSpeed != dbc.FanSpeed {
		updates["FanSpeed"] = ic.FanSpeed
	}
	if ic.VentVertical != dbc.VentVertical {
		updates["VentVertical"] = ic.VentVertical
	}
	if ic.VentHorizontal != dbc.VentHorizontal {
		updates["VentHorizontal"] = ic.VentHorizontal
	}
	if ic.TimerOnEnabled != dbc.TimerOnEnabled {
		updates["TimerOnEnabled"] = ic.TimerOnEnabled
	}
	if ic.TimerOffEnabled != dbc.TimerOffEnabled {
		updates["TimerOffEnabled"] = ic.TimerOffEnabled
	}
	if ic.TimerOn != dbc.TimerOn && ic.TimerOn != codec.C_Time_Unset {
		updates["TimerOn"] = ic.TimerOn
	}
	if ic.TimerOff != dbc.TimerOff && ic.TimerOff != codec.C_Time_Unset {
		updates["TimerOff"] = ic.TimerOff
	}

	var nc DbIrConfig
	myDb.First(&nc, 1)
	result := myDb.Model(&nc).Updates(updates)
	if result.Error != nil {
		return result.Error
	}

	settings["Temperature"] = ic.Temperature
	if ic.Powerful == codec.C_Powerful_Disabled && ic.Quiet == codec.C_Quiet_Disabled {
		settings["FanSpeed"] = ic.FanSpeed
	}
	// fmt.Printf("updating settings for mode=%d: ", ic.Mode)
	// fmt.Println(settings)
	result = myDb.Model(ModeSetting{}).Where("mode = ?", ic.Mode).Updates(settings)
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
	// fmt.Printf("getting settings for mode=%d: temp=%d fan=%d\n", ms.Mode, ms.Temperature, ms.FanSpeed)
	return ms.Temperature, ms.FanSpeed, nil
}
