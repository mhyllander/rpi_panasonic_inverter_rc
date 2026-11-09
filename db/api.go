package db

import (
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
)

var myDb *gorm.DB

// Serializes all read-modify-write sequences on the current config, so that
// concurrent callers (web, scheduler, IR receiver) can't overwrite each other.
var configMu sync.Mutex

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
		return fmt.Errorf("failed to open database %q: %w", dbFile, err)
	}

	// Migrate the schema
	err = myDb.AutoMigrate(&DbIrConfig{}, &ModeSetting{}, &JobSet{}, &CronJob{})
	if err != nil {
		return fmt.Errorf("failed to migrate database schema: %w", err)
	}

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
			return result.Error
		}

		settings := []*ModeSetting{
			{Mode: codecbase.C_Mode_Auto, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
			{Mode: codecbase.C_Mode_Dry, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
			{Mode: codecbase.C_Mode_Heat, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
			{Mode: codecbase.C_Mode_Cool, Temperature: 20, FanSpeed: codecbase.C_FanSpeed_Auto},
		}
		result = myDb.Create(settings)
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func Close() {

}

func CurrentConfig() (*codec.RcConfig, error) {
	return currentConfig(myDb)
}

func currentConfig(conn *gorm.DB) (*codec.RcConfig, error) {
	var dbRc DbIrConfig
	result := conn.First(&dbRc, 1)
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

// WithConfig calls fn with the current config while holding the config lock, so
// that nothing can update the config until fn returns.
func WithConfig(fn func(current *codec.RcConfig) error) error {
	configMu.Lock()
	defer configMu.Unlock()

	current, err := currentConfig(myDb)
	if err != nil {
		return err
	}
	return fn(current)
}

// UpdateConfig atomically reads the current config, lets compose produce the new
// config, and saves it in a transaction. If compose returns a nil config, nothing
// is saved. The optional afterSave is called with the resulting config (the
// current config if nothing was saved) after a successful commit, still holding
// the config lock, so that anything done in it (e.g. sending IR) is ordered the
// same way as the saved updates. If afterSave fails, the config has still been saved.
func UpdateConfig(compose func(current *codec.RcConfig) (*codec.RcConfig, error), afterSave func(result *codec.RcConfig) error) error {
	configMu.Lock()
	defer configMu.Unlock()

	var result *codec.RcConfig
	err := myDb.Transaction(func(tx *gorm.DB) error {
		current, err := currentConfig(tx)
		if err != nil {
			return err
		}

		newConfig, err := compose(current)
		if err != nil {
			return err
		}
		if newConfig == nil {
			result = current
			return nil
		}

		result = newConfig
		return saveConfig(tx, newConfig, current)
	})
	if err != nil {
		return err
	}

	if afterSave != nil {
		return afterSave(result)
	}
	return nil
}

// SaveConfig saves rc without holding the lock across the caller's read. It is
// only safe for single-threaded callers such as the command line utility; the
// controller must use UpdateConfig.
func SaveConfig(rc, dbRc *codec.RcConfig) error {
	configMu.Lock()
	defer configMu.Unlock()
	return saveConfig(myDb, rc, dbRc)
}

func saveConfig(conn *gorm.DB, rc, dbRc *codec.RcConfig) error {
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
	if result := conn.First(&nc, 1); result.Error != nil {
		return result.Error
	}
	if result := conn.Model(&nc).Updates(updates); result.Error != nil {
		return result.Error
	}

	settings := make(map[string]interface{})
	settings["Temperature"] = rc.Temperature
	if rc.Powerful == codecbase.C_Powerful_Disabled && rc.Quiet == codecbase.C_Quiet_Disabled {
		settings["FanSpeed"] = rc.FanSpeed
	}
	if result := conn.Model(&ModeSetting{}).Where(map[string]interface{}{"Mode": rc.Mode}).Updates(settings); result.Error != nil {
		return result.Error
	}

	slog.Debug("saved config to db")
	return nil
}

func SetPower(power uint) error {
	configMu.Lock()
	defer configMu.Unlock()

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

func saveCronJob(conn *gorm.DB, jobset, name, schedule string, settings *codecbase.Settings) error {
	settingsJson, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	cj := CronJob{JobSet: jobset, Name: name, Schedule: schedule, Settings: settingsJson}
	return conn.Create(&cj).Error
}

func GetCronJobs(jobset string) (*[]CronJob, error) {
	var cronjobs []CronJob
	if result := myDb.Where(&CronJob{JobSet: jobset}).Order("id").Find(&cronjobs); result.Error != nil {
		return nil, result.Error
	}
	return &cronjobs, nil
}

func DeleteAllCronJobsPermanently() error {
	return deleteAllCronJobsPermanently(myDb)
}

func deleteAllCronJobsPermanently(conn *gorm.DB) error {
	// AllowGlobalUpdate needed to delete all, Unscoped needed to bypass soft delete
	return conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&CronJob{}).Error
}

// JobSets

func SaveJobSet(jobset string, active bool) error {
	return saveJobSet(myDb, jobset, active)
}

func saveJobSet(conn *gorm.DB, jobset string, active bool) error {
	js := JobSet{Name: jobset, Active: active}
	return conn.Create(&js).Error
}

func GetJobSets() (*[]JobSet, error) {
	var jobsets []JobSet
	if result := myDb.Find(&jobsets); result.Error != nil {
		return nil, result.Error
	}
	return &jobsets, nil
}

func UpdateJobSet(jobset string, active bool) error {
	return myDb.Model(&JobSet{}).Where("name = ?", jobset).Updates(map[string]interface{}{"Active": active}).Error
}

func GetActiveJobSets() (*[]JobSet, error) {
	var jobsets []JobSet
	if result := myDb.Where(map[string]interface{}{"Active": true}).Find(&jobsets); result.Error != nil {
		return nil, result.Error
	}
	return &jobsets, nil
}

func DeleteAllJobSetsPermanently() error {
	return deleteAllJobSetsPermanently(myDb)
}

func deleteAllJobSetsPermanently(conn *gorm.DB) error {
	// AllowGlobalUpdate needed to delete all, Unscoped needed to bypass soft delete
	return conn.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(&JobSet{}).Error
}

// ReplaceJobSets replaces all job sets and cron jobs in a single transaction.
// If anything fails, the existing job sets and cron jobs are left untouched.
func ReplaceJobSets(definitions []JobSetDefinition) error {
	return myDb.Transaction(func(tx *gorm.DB) error {
		if err := deleteAllJobSetsPermanently(tx); err != nil {
			return fmt.Errorf("failed to delete jobsets: %w", err)
		}
		if err := deleteAllCronJobsPermanently(tx); err != nil {
			return fmt.Errorf("failed to delete cronjobs: %w", err)
		}

		for _, definition := range definitions {
			if err := saveJobSet(tx, definition.Name, definition.Active); err != nil {
				return fmt.Errorf("failed to save jobset %q: %w", definition.Name, err)
			}
			for _, cronJob := range definition.CronJobs {
				if err := saveCronJob(tx, definition.Name, cronJob.Name, cronJob.Schedule, &cronJob.Settings); err != nil {
					return fmt.Errorf("failed to save cronjob %q in jobset %q: %w", cronJob.Schedule, definition.Name, err)
				}
			}
		}
		return nil
	})
}
