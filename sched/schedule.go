package sched

import (
	"encoding/json"
	"log/slog"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/common"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/utils"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var scheduler gocron.Scheduler
var g_irSender *codec.IrSender

func SendCurrentConfig() {
	slog.Info("initializing: sending current config")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}

	sendRc := dbRc.CopyForSendingAll()
	// update power setting, adjusting for any timers
	utils.SetPower("", sendRc, dbRc)

	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("failed to save config", "err", err)
		return
	}
	slog.Debug("saved config")
}

func RunCronJob(settings *common.Settings) {
	slog.Info("processing cronjob")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}

	sendRc := utils.ComposeSendConfig(settings, dbRc)
	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("failed to save config", "err", err)
		return
	}
	slog.Debug("saved config")
}

func createCronJobs() {
	if jss, err := db.GetActiveJobSets(); err != nil {
		slog.Error("failed to get active jobsets", "err", err)
	} else {
		for _, js := range *jss {
			slog.Info("Scheduling job set", "jobset", js.Name)
			cjs, err := db.GetCronJobs(js.Name)
			if err != nil {
				slog.Error("failed to get cronjobs", "err", err)
				continue
			}

			for _, cj := range *cjs {
				var settings common.Settings
				err = json.Unmarshal(cj.Settings, &settings)
				if err != nil {
					slog.Error("failed to unmarshal json", "err", err)
					break
				}
				j, err := scheduler.NewJob(
					gocron.CronJob(cj.Schedule, false),
					gocron.NewTask(
						RunCronJob,
						&settings,
					),
				)
				if err != nil {
					slog.Error("failed to schedule cronjob", "schedule", cj.Schedule, "err", err)
					break
				}
				slog.Debug("scheduled cronjob", "job_id", j.ID())
			}
		}
	}
}

func InitScheduler(irSender *codec.IrSender) error {
	var err error

	g_irSender = irSender

	// create a scheduler
	scheduler, err = gocron.NewScheduler(
		gocron.WithLogger(slog.Default()),
		gocron.WithLimitConcurrentJobs(1, gocron.LimitModeWait),
	)
	if err != nil {
		return err
	}

	// Send the current config one minute after start. This is so that the inverter will
	// be configured in case of a power outage, in which case both the RPi and the inverter
	// will be restarted.
	j, err := scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(2*time.Minute)),
		),
		gocron.NewTask(
			SendCurrentConfig,
		),
	)
	if err != nil {
		return err
	}
	slog.Debug("scheduled initial job", "job_id", j.ID())

	// Schedule all the active cron jobs
	createCronJobs()

	// start the scheduler
	scheduler.Start()

	return nil
}

func Stop() {
	// when you're done, shut it down
	err := scheduler.Shutdown()
	if err != nil {
		slog.Error("failed to stop scheduler", "error", err)
	}
}
