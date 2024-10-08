package db

import (
	"encoding/json"
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
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
	myDb.AutoMigrate(&DbIrConfig{}, &ModeSetting{}, &JobSet{}, &CronJob{})

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
			{Mode: codecbase.C_Mode_Auto, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
			{Mode: codecbase.C_Mode_Dry, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
			{Mode: codecbase.C_Mode_Heat, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
			{Mode: codecbase.C_Mode_Cool, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
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
		Clock:          codecbase.C_Time_Unset,
	}, nil
}

func SaveConfig(rc, dbRc *codec.RcConfig) error {
	// update current configuration, but timer on and off should only be updated if set
	// mode settings should be updated, but fan speed should be ignored if Powerful or Quiet is set

	updates := make(map[string]interface{})
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
	if rc.TimerOnTime != dbRc.TimerOnTime && rc.TimerOnTime != codecbase.C_Time_Unset {
		updates["TimerOnTime"] = rc.TimerOnTime
	}
	if rc.TimerOffTime != dbRc.TimerOffTime && rc.TimerOffTime != codecbase.C_Time_Unset {
		updates["TimerOffTime"] = rc.TimerOffTime
	}

	var nc DbIrConfig
	if result := myDb.First(&nc, 1); result.Error != nil {
		return result.Error
	}
	if result := myDb.Model(&nc).Updates(updates); result.Error != nil {
		return result.Error
	}

	settings := make(map[string]interface{})
	settings["Temperature"] = rc.Temperature
	if rc.Powerful == codecbase.C_Powerful_Disabled && rc.Quiet == codecbase.C_Quiet_Disabled {
		settings["FanSpeed"] = rc.FanSpeed
	}
	if result := myDb.Model(&ModeSetting{}).Where(map[string]interface{}{"Mode": rc.Mode}).Updates(settings); result.Error != nil {
		return result.Error
	}

	slog.Debug("saved config to db")
	return nil
}

func SetPower(power uint) error {
	var nc DbIrConfig
	if result := myDb.First(&nc, 1); result.Error != nil {
		return result.Error
	}
	if result := myDb.Model(&nc).Updates(map[string]interface{}{"Power": power}); result.Error != nil {
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

// CronJob

func SaveCronJob(jobset string, schedule string, settings *codecbase.Settings) error {
	json, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	cj := CronJob{JobSet: jobset, Schedule: schedule, Settings: json}
	myDb.Create(&cj)
	return nil
}

func GetCronJobs(jobset string) (*[]CronJob, error) {
	var cronjobs []CronJob
	if result := myDb.Where(&CronJob{JobSet: jobset}).Find(&cronjobs); result.Error != nil {
		return nil, result.Error
	}
	return &cronjobs, nil
}

func DeleteAllCronJobsPermanently() {
	// AllowGlobalUpdate needed to delete all, Unscoped needed to bypass soft delete
	myDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&CronJob{})
}

// JobSets

func SaveJobSet(jobset string, active bool) error {
	cj := JobSet{Name: jobset, Active: active}
	myDb.Create(&cj)
	return nil
}

func GetJobSets() (*[]JobSet, error) {
	var jobsets []JobSet
	if result := myDb.Find(&jobsets); result.Error != nil {
		return nil, result.Error
	}
	return &jobsets, nil
}

func UpdateJobSet(jobset string, active bool) error {
	myDb.Model(&JobSet{}).Where("name = ?", jobset).Updates(map[string]interface{}{"Active": active})
	return nil
}

func GetActiveJobSets() (*[]JobSet, error) {
	var jobsets []JobSet
	if result := myDb.Where(map[string]interface{}{"Active": true}).Find(&jobsets); result.Error != nil {
		return nil, result.Error
	}
	return &jobsets, nil
}

func DeleteAllJobSetsPermanently() {
	// AllowGlobalUpdate needed to delete all, Unscoped needed to bypass soft delete
	myDb.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&JobSet{})
}
