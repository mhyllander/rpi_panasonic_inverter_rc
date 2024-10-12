package sched

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-co-op/gocron/v2"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/logs"
	"rpi_panasonic_inverter_rc/rcutils"
)

var scheduler gocron.Scheduler
var g_irSender *codec.IrSender

const settingsJobCategory = "settings"
const timerJobCategory = "timer"

func RunInitializationJob() {
	slog.Info("running initialization job")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("RunInitializationJob: failed to get current config", "err", err)
		return
	}

	sendRc := dbRc.CopyForSendingAll()
	g_irSender.SendConfig(sendRc)
}

func RunSettingsJob(settings codecbase.Settings, jobName string) {
	slog.Info("running settings job", "jobName", jobName)

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("RunSettingsJob: failed to get current config", "err", err)
		return
	}

	sendRc := rcutils.ComposeSendConfig(&settings, dbRc)
	g_irSender.SendConfig(sendRc)

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("RunSettingsJob: failed to save config", "err", err)
		return
	}

	CondRestartTimerJobs(&settings)
}

func ScheduleJobsForJobset(jobset string, active bool) {
	// remove existing jobs of the current generation
	jobsetGen := jobsetGens.currentGen(settingsJobCategory, jobset)
	scheduler.RemoveByTags(jobsetGen)

	if !active {
		slog.Info("Unscheduled inactive jobset", "jobset", jobset, "jobsetGen", jobsetGen)
		return
	}

	// get the next job generation before creating new jobs
	jobsetGen = jobsetGens.nextGen(settingsJobCategory, jobset)
	slog.Info("Scheduling jobset", "jobset", jobset, "jobsetGen", jobsetGen)

	cjs, err := db.GetCronJobs(jobset)
	if err != nil {
		slog.Error("failed to get cronjobs", "err", err)
		return
	}

	for _, cj := range *cjs {
		var settings = new(codecbase.Settings)
		err = json.Unmarshal(cj.Settings, settings)
		if err != nil {
			slog.Error("failed to unmarshal json", "err", err)
			break
		}
		name := fmt.Sprintf("%s_%d %s", jobset, cj.ID, cj.Schedule)
		j, err := scheduler.NewJob(
			gocron.CronJob(cj.Schedule, false),
			gocron.NewTask(
				RunSettingsJob,
				*settings,
				name,
			),
			gocron.WithName(name),
			gocron.WithTags(settingsJobCategory, jobset, jobsetGen),
		)
		if err != nil {
			slog.Error("failed to schedule settings job", "schedule", cj.Schedule, "err", err)
			break
		}
		slog.Debug("scheduled settings job", "jobset", jobset, "jobsetGen", jobsetGen, "job_id", j.ID())
	}
	listJobs()
}

func createSettingsJobs() {
	if jss, err := db.GetActiveJobSets(); err != nil {
		slog.Error("failed to get active jobsets", "err", err)
	} else {
		for _, js := range *jss {
			ScheduleJobsForJobset(js.Name, js.Active)
		}
	}
}

func RunTimerJob(power uint, jobName string) {
	slog.Info("running timer job", "jobName", jobName, "power", power)
	if err := db.SetPower(power); err != nil {
		slog.Error("RunTimerJob: failed to set power", "err", err)
		return
	}
	slog.Debug("RunTimerJob: updated power", "power", power)
}

func scheduleTimerJob(jobName, jobsetGen string, power uint, t codec.Time) (gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(t.Hour(), t.Minute(), 0))),
		gocron.NewTask(
			RunTimerJob,
			power,
			jobName,
		),
		gocron.WithName(jobName),
		gocron.WithTags(timerJobCategory, jobsetGen),
	)
	if err != nil {
		slog.Error("failed to schedule timer job", "jobName", jobName, "jobsetGen", jobsetGen, "at", t.ToString(), "err", err)
		return nil, err
	}
	slog.Info("scheduled timer job", "jobName", jobName, "jobsetGen", jobsetGen, "at", t.ToString())
	return j, nil
}

func RunDstTransitionJob(jobName, jobsetGen string) {
	slog.Info("running DST transition job", "jobName", jobName, "jobsetGen", jobsetGen)

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("RunDstTransitionJob: failed to get current config", "err", err)
		return
	}

	// re-send the current configuration with an updated clock
	sendRc := dbRc.CopyForSendingAll()
	g_irSender.SendConfig(sendRc)

	// re-schedule the next DST transition job
	scheduleDstTransitionJob(jobName, jobsetGen)
}

func scheduleDstTransitionJob(jobName, jobsetGen string) (gocron.Job, error) {
	_, end := time.Now().ZoneBounds()
	j, err := scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(end)),
		gocron.NewTask(
			RunDstTransitionJob,
			jobName,
			jobsetGen,
		),
		gocron.WithName(jobName),
		gocron.WithTags(timerJobCategory, jobsetGen),
	)
	if err != nil {
		slog.Error("failed to schedule DST transition job", "jobName", jobName, "jobsetGen", jobsetGen, "at", end, "err", err)
		return nil, err
	}
	slog.Info("scheduled DST transition job", "jobName", jobName, "jobsetGen", jobsetGen, "at", end)
	return j, nil
}

func CondRestartTimerJobs(settings *codecbase.Settings) {
	if settings.TimerOn != "" || settings.TimerOnTime != "" ||
		settings.TimerOff != "" || settings.TimerOffTime != "" {
		RestartTimerJobs()
	}
}

// The inverter can power on up to 60 minutes before the configured time to achieve the desired temperature, depending
// on the current indoor and outdoor temperatures (it actually turns on silently to check the temperatures so it can
// decide how long time before to power on). After being off during the night it sometimes starts 45 minutes before.
// Since we can't know when it will start, we'll create two timer on jobs, one that runs 60 minutes before time and one
// that runs on time.
func RestartTimerJobs() {
	slog.Debug("restarting timer jobs")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}

	// remove existing jobs of the current generation
	jobsetGen := jobsetGens.currentGen(timerJobCategory, timerJobCategory)
	scheduler.RemoveByTags(jobsetGen)

	// get the next job generation before creating new jobs
	jobsetGen = jobsetGens.nextGen(timerJobCategory, timerJobCategory)

	const job1MinutesBefore = 60
	if dbRc.TimerOn == codecbase.C_Timer_Enabled {
		job1Time := dbRc.TimerOnTime - job1MinutesBefore
		scheduleTimerJob("timer_on_1", jobsetGen, codecbase.C_Power_On, job1Time)
		job2Time := dbRc.TimerOnTime
		scheduleTimerJob("timer_on_2", jobsetGen, codecbase.C_Power_On, job2Time)
	}
	if dbRc.TimerOff == codecbase.C_Timer_Enabled {
		jobTime := dbRc.TimerOffTime
		scheduleTimerJob("timer_off", jobsetGen, codecbase.C_Power_Off, jobTime)
	}

	// schedule a DST transition job if timers are enabled
	if dbRc.TimerOn == codecbase.C_Timer_Enabled || dbRc.TimerOff == codecbase.C_Timer_Enabled {
		scheduleDstTransitionJob("dst_transition", jobsetGen)
	}

	listJobs()
	slog.Debug("updated timer jobs")
}

func listJobs() {
	if logs.IsLogLevelDebug() {
		for _, j := range scheduler.Jobs() {
			slog.Debug("listJobs", "id", j.ID(), "name", j.Name(), "tags", j.Tags())
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

	// Send the current config after start. This ensures that the inverter will be re-configured after a power outage.
	_, err = scheduler.NewJob(
		gocron.OneTimeJob(
			gocron.OneTimeJobStartDateTime(time.Now().Add(2*time.Minute)),
		),
		gocron.NewTask(
			RunInitializationJob,
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
