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

	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("failed to save config", "err", err)
		return
	}
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

func TimerJob(power uint) {
	slog.Info("processing timer job")
	if err := db.SetPower(power); err != nil {
		slog.Error("TimerJob: failed to set power", "err", err)
		return
	}
	slog.Debug("TimerJob: updated power", "power", power)
}

func scheduleTimerJob(name string, power uint, t codec.Time) (gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(t.Hour(), t.Minute(), 0))),
		gocron.NewTask(
			TimerJob,
			power,
		),
		gocron.WithName(name),
	)
	if err != nil {
		return nil, err
	}
	return j, nil
}

var timerOnJob, timerOffJob gocron.Job

func UpdateTimerJobs() {
	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("UpdateTimerJobs: failed to get current config", "err", err)
		return
	}

	if timerOnJob != nil {
		if err := scheduler.RemoveJob(timerOnJob.ID()); err != nil {
			slog.Error("UpdateTimerJobs: failed to remove timer on job", "err", err)
		}
		timerOnJob = nil
	}
	if dbRc.TimerOn == common.C_Timer_Enabled {
		// The inverter can power on before the configured time to achieve the desired temperature. After being off
		// during the night it sometimes starts 45 minutes before. It probably depends on the difference between the
		// current and the desired temperatures. Since we can't know when it will start, we'll just have to use a
		// hard-coded default.
		t := dbRc.TimerOnTime - 45
		if j, err := scheduleTimerJob("timer_on", common.C_Power_On, t); err == nil {
			timerOnJob = j
			slog.Info("UpdateTimerJobs: scheduled timer on job", "t", t.ToString())
		} else {
			slog.Error("UpdateTimerJobs: failed to schedule timer on job", "err", err)
		}
	}

	if timerOffJob != nil {
		if err := scheduler.RemoveJob(timerOffJob.ID()); err != nil {
			slog.Error("UpdateTimerJobs: failed to remove timer off job", "err", err)
		}
		timerOffJob = nil
	}
	if dbRc.TimerOff == common.C_Timer_Enabled {
		if j, err := scheduleTimerJob("timer_off", common.C_Power_Off, dbRc.TimerOffTime); err == nil {
			timerOffJob = j
			slog.Info("UpdateTimerJobs: scheduled timer off job", "t", dbRc.TimerOffTime.ToString())
		} else {
			slog.Error("UpdateTimerJobs: failed to schedule timer off job", "err", err)
		}
	}

	slog.Debug("updated timer jobs")
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
	UpdateTimerJobs()

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
