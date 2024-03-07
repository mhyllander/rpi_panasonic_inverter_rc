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

const settingsJobTag = "settings"
const timerJobTag = "timer"

func SendCurrentConfig() {
	slog.Info("running initialization job: send current config")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("SendCurrentConfig: failed to get current config", "err", err)
		return
	}

	sendRc := dbRc.CopyForSendingAll()
	g_irSender.SendConfig(sendRc)
}

func RunSettingsJob(settings *common.Settings) {
	slog.Info("running settings job")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("RunSettingsJob: failed to get current config", "err", err)
		return
	}

	sendRc := utils.ComposeSendConfig(settings, dbRc)
	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("RunSettingsJob: failed to save config", "err", err)
		return
	}

	CondRestartTimerJobs(settings)
}

func createSettingsJobs() {
	if jss, err := db.GetActiveJobSets(); err != nil {
		slog.Error("failed to get active jobsets", "err", err)
	} else {
		for _, js := range *jss {
			slog.Info("Scheduling jobset", "jobset", js.Name)
			cjs, err := db.GetCronJobs(js.Name)
			if err != nil {
				slog.Error("failed to get cronjobs", "err", err)
				continue
			}

			for _, cj := range *cjs {
				var settings = new(common.Settings)
				err = json.Unmarshal(cj.Settings, settings)
				if err != nil {
					slog.Error("failed to unmarshal json", "err", err)
					break
				}
				j, err := scheduler.NewJob(
					gocron.CronJob(cj.Schedule, false),
					gocron.NewTask(
						RunSettingsJob,
						settings,
					),
					gocron.WithTags(settingsJobTag, js.Name),
				)
				if err != nil {
					slog.Error("failed to schedule settings job", "schedule", cj.Schedule, "err", err)
					break
				}
				slog.Debug("scheduled settings job", "js", js.Name, "job_id", j.ID())
			}
		}
	}
}

func RunTimerJob(power uint, jobName string) {
	slog.Info("running timer job", "job", jobName, "power", power)
	if err := db.SetPower(power); err != nil {
		slog.Error("RunTimerJob: failed to set power", "err", err)
		return
	}
	slog.Debug("RunTimerJob: updated power", "power", power)
}

func scheduleTimerJob(jobName string, power uint, t codec.Time) (gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(t.Hour(), t.Minute(), 0))),
		gocron.NewTask(
			RunTimerJob,
			power,
			jobName,
		),
		gocron.WithName(jobName),
		gocron.WithTags(timerJobTag),
	)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func CondRestartTimerJobs(settings *common.Settings) {
	if settings.TimerOn != "" || settings.TimerOnTime != "" ||
		settings.TimerOff != "" || settings.TimerOffTime != "" {
		RestartTimerJobs()
	}
}

// The inverter can power on before the configured time to achieve the desired temperature. After being off during
// the night it sometimes starts 45 minutes before. It probably depends on the difference between the current and
// the desired temperatures. Since we can't know when it will start, we'll create two timer on jobs, one that runs
// 45 minutes before time and one that runs on time.
func RestartTimerJobs() {
	slog.Debug("restarting timer jobs")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}

	// If you remove two jobs individually directly after one another, the second removal fails with "job not found".
	// But removing multiple jobs by tag works fine.
	scheduler.RemoveByTags(timerJobTag)

	const job1MinutesBefore = 45
	if dbRc.TimerOn == common.C_Timer_Enabled {
		job1Time := dbRc.TimerOnTime - job1MinutesBefore
		if _, err := scheduleTimerJob("timer_on_1", common.C_Power_On, job1Time); err == nil {
			slog.Info("scheduled first timer_on job", "at", job1Time.ToString())
		} else {
			slog.Error("failed to schedule first timer_on job", "err", err)
		}
		job2Time := dbRc.TimerOnTime
		if _, err := scheduleTimerJob("timer_on_2", common.C_Power_On, job2Time); err == nil {
			slog.Info("scheduled second timer_on job", "at", job2Time.ToString())
		} else {
			slog.Error("failed to schedule second timer_on job", "err", err)
		}
	}
	if dbRc.TimerOff == common.C_Timer_Enabled {
		jobTime := dbRc.TimerOffTime
		if _, err := scheduleTimerJob("timer_off", common.C_Power_Off, jobTime); err == nil {
			slog.Info("scheduled timer_off job", "at", jobTime.ToString())
		} else {
			slog.Error("failed to schedule timer_off job", "err", err)
		}
	}

	listJobs()
	slog.Debug("updated timer jobs")
}

func listJobs() {
	if common.IsLogLevelDebug() {
		for _, j := range scheduler.Jobs() {
			slog.Debug("job", "id", j.ID(), "name", j.Name(), "tags", j.Tags())
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

	// start the scheduler
	scheduler.Start()

	// Send the current config after start. This ensures that the inverter will be configured after a power outage.
	_, err = scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(2*time.Minute)),
		),
		gocron.NewTask(
			SendCurrentConfig,
		),
		gocron.WithName("initialization"),
	)
	if err != nil {
		return err
	}
	slog.Info("Scheduled initialization job")

	// Schedule all the active cron jobs
	createSettingsJobs()
	RestartTimerJobs()

	return nil
}

func Stop() {
	// when you're done, shut it down
	err := scheduler.Shutdown()
	if err != nil {
		slog.Error("failed to stop scheduler", "error", err)
	}
}
