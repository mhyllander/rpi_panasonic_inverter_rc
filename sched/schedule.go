package sched

import (
	"encoding/json/v2"
	"errors"
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

const minutesPerDay = 24 * 60

func RunInitializationJob() {
	slog.Info("running initialization job")

	err := db.WithConfig(func(current *codec.RcConfig) error {
		return g_irSender.SendConfig(current.CopyForSendingAll())
	})
	if err != nil {
		slog.Error("RunInitializationJob: failed to send current config", "err", err)
	}
}

func RunSettingsJob(settings codecbase.Settings, jobName string) {
	slog.Info("running settings job", "jobName", jobName)

	err := db.UpdateConfig(
		func(current *codec.RcConfig) (*codec.RcConfig, error) {
			return rcutils.ComposeSendConfig(&settings, current), nil
		},
		g_irSender.SendConfig,
	)
	if err != nil {
		slog.Error("RunSettingsJob: failed to update config", "err", err)
		return
	}

	if err := CondRestartTimerJobs(&settings); err != nil {
		slog.Error("RunSettingsJob: failed to restart timer jobs", "err", err)
	}
}

// ScheduleJobsForJobset replaces the scheduled jobs of a jobset. A job that can't be scheduled is skipped so
// that it doesn't prevent the other jobs in the jobset from running. The returned error lists all skipped jobs.
func ScheduleJobsForJobset(jobset string, active bool) error {
	// remove existing jobs of the current generation
	jobsetGen := jobsetGens.currentGen(settingsJobCategory, jobset)
	scheduler.RemoveByTags(jobsetGen)

	if !active {
		slog.Info("Unscheduled inactive jobset", "jobset", jobset, "jobsetGen", jobsetGen)
		return nil
	}

	// get the next job generation before creating new jobs
	jobsetGen = jobsetGens.nextGen(settingsJobCategory, jobset)
	slog.Info("Scheduling jobset", "jobset", jobset, "jobsetGen", jobsetGen)

	cjs, err := db.GetCronJobs(jobset)
	if err != nil {
		return fmt.Errorf("failed to get cronjobs for jobset %q: %w", jobset, err)
	}

	var skippedJobs []error
	for _, cj := range *cjs {
		var settings = new(codecbase.Settings)
		if err := json.Unmarshal(cj.Settings, settings); err != nil {
			skippedJobs = append(skippedJobs, fmt.Errorf("cronjob %d in jobset %q has invalid settings: %w", cj.ID, jobset, err))
			continue
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
			skippedJobs = append(skippedJobs, fmt.Errorf("cronjob %d in jobset %q with schedule %q could not be scheduled: %w", cj.ID, jobset, cj.Schedule, err))
			continue
		}
		slog.Debug("scheduled settings job", "jobset", jobset, "jobsetGen", jobsetGen, "job_id", j.ID())
	}
	listJobs()

	return errors.Join(skippedJobs...)
}

func createSettingsJobs() {
	if jss, err := db.GetActiveJobSets(); err != nil {
		slog.Error("failed to get active jobsets", "err", err)
	} else {
		for _, js := range *jss {
			if err := ScheduleJobsForJobset(js.Name, js.Active); err != nil {
				slog.Error("failed to schedule all jobs in jobset", "jobset", js.Name, "err", err)
			}
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
	if t.Minutes() >= minutesPerDay {
		err := fmt.Errorf("time %s is not a valid time of day", t.ToString())
		slog.Error("failed to schedule timer job", "jobName", jobName, "jobsetGen", jobsetGen, "err", err)
		return nil, err
	}

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

	// re-send the current configuration with an updated clock
	err := db.WithConfig(func(current *codec.RcConfig) error {
		return g_irSender.SendConfig(current.CopyForSendingAll())
	})
	if err != nil {
		slog.Error("RunDstTransitionJob: failed to send current config", "err", err)
	}

	// re-schedule the next DST transition job
	if _, err := scheduleDstTransitionJob(jobName, jobsetGen); err != nil {
		slog.Error("RunDstTransitionJob: failed to schedule the next DST transition job", "err", err)
	}
}

// Schedules a job at the next DST transition. Returns a nil job if the time zone has no upcoming transition.
func scheduleDstTransitionJob(jobName, jobsetGen string) (gocron.Job, error) {
	now := time.Now()
	_, end := now.ZoneBounds()
	if end.IsZero() {
		// Scheduling a job at the zero time would run it immediately, and since the job schedules
		// itself again it would keep running in a loop.
		slog.Info("no upcoming DST transition in the current time zone, not scheduling DST transition job", "jobName", jobName)
		return nil, nil
	}
	if !end.After(now) {
		return nil, fmt.Errorf("next DST transition at %s is not in the future", end)
	}
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

func CondRestartTimerJobs(settings *codecbase.Settings) error {
	if settings.TimerOn != "" || settings.TimerOnTime != "" ||
		settings.TimerOff != "" || settings.TimerOffTime != "" {
		return RestartTimerJobs()
	}
	return nil
}

// The inverter can power on up to 60 minutes before the configured time to achieve the desired temperature, depending
// on the current indoor and outdoor temperatures (it actually turns on silently to check the temperatures so it can
// decide how long time before to power on). After being off during the night it sometimes starts 45 minutes before.
// Since we can't know when it will start, we'll create two timer on jobs, one that runs 60 minutes before time and one
// that runs on time.
func RestartTimerJobs() error {
	slog.Debug("restarting timer jobs")

	dbRc, err := db.CurrentConfig()
	if err != nil {
		return fmt.Errorf("failed to get current config: %w", err)
	}

	// remove existing jobs of the current generation
	jobsetGen := jobsetGens.currentGen(timerJobCategory, timerJobCategory)
	scheduler.RemoveByTags(jobsetGen)

	// get the next job generation before creating new jobs
	jobsetGen = jobsetGens.nextGen(timerJobCategory, timerJobCategory)

	// Try to schedule all jobs even if some fail. The error is logged where the job is scheduled.
	var failedJobs []error
	collectError := func(_ gocron.Job, err error) {
		if err != nil {
			failedJobs = append(failedJobs, err)
		}
	}

	const job1MinutesBefore = 60
	if dbRc.TimerOn == codecbase.C_Timer_Enabled {
		var onTime, preOnTime codec.Time
		onTime = dbRc.TimerOnTime
		if onTime >= job1MinutesBefore {
			preOnTime = onTime - job1MinutesBefore
		} else {
			preOnTime = minutesPerDay + onTime - job1MinutesBefore
		}
		collectError(scheduleTimerJob("timer_on_pre", jobsetGen, codecbase.C_Power_On, preOnTime))
		collectError(scheduleTimerJob("timer_on", jobsetGen, codecbase.C_Power_On, onTime))
	}
	if dbRc.TimerOff == codecbase.C_Timer_Enabled {
		jobTime := dbRc.TimerOffTime
		collectError(scheduleTimerJob("timer_off", jobsetGen, codecbase.C_Power_Off, jobTime))
	}

	// schedule a DST transition job if timers are enabled
	if dbRc.TimerOn == codecbase.C_Timer_Enabled || dbRc.TimerOff == codecbase.C_Timer_Enabled {
		collectError(scheduleDstTransitionJob("dst_transition", jobsetGen))
	}

	listJobs()
	slog.Debug("updated timer jobs")

	return errors.Join(failedJobs...)
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
	if err := RestartTimerJobs(); err != nil {
		slog.Error("failed to schedule all timer jobs", "err", err)
	}

	return nil
}

func Stop() {
	// when you're done, shut it down
	err := scheduler.Shutdown()
	if err != nil {
		slog.Error("failed to stop scheduler", "error", err)
	}
}
