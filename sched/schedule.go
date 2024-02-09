package sched

import (
	"encoding/json"
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/rcconst"
	"rpi_panasonic_inverter_rc/utils"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var scheduler gocron.Scheduler
var g_senderOptions *codec.SenderOptions
var g_irOutputFile string

func openIrOutputFile() *os.File {
	// open file or device for sending IR
	flags := os.O_RDWR
	if !g_senderOptions.Device {
		flags = flags | os.O_CREATE
	}
	f, err := os.OpenFile(g_irOutputFile, flags, 0644)
	if err != nil {
		slog.Error("failed to open IR output file", "err", err)
		return nil
	}
	return f
}

func SendCurrentConfig() {
	slog.Info("initializing: sending current config")

	confirmCommand := make(chan struct{})
	codec.SuspendReceiver(confirmCommand)
	<-confirmCommand
	defer func() {
		codec.ResumeReceiver(confirmCommand)
		<-confirmCommand
	}()

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}
	sendRc := dbRc.CopyForSendingAll()
	// update power setting, adjusting for any timers
	utils.SetPower("", sendRc, dbRc)

	f := openIrOutputFile()
	if f == nil {
		slog.Error("failed to open IR output file", "err", err)
		return
	}
	defer f.Close()

	sendRc.LogConfigAndChecksum("")
	err = codec.SendIr(sendRc, f, g_senderOptions)
	if err != nil {
		slog.Error("failed to send current config", "err", err)
		return
	}

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		slog.Error("failed to save config", "err", err)
		return
	}
	slog.Debug("saved config")
}

func RunCronJob(settings *rcconst.Settings) {
	slog.Info("processing cronjob")

	confirmCommand := make(chan struct{})
	codec.SuspendReceiver(confirmCommand)
	<-confirmCommand
	defer func() {
		codec.ResumeReceiver(confirmCommand)
		<-confirmCommand
	}()

	dbRc, err := db.CurrentConfig()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}

	sendRc := utils.ComposeSendConfig(settings, dbRc)

	f := openIrOutputFile()
	if f == nil {
		slog.Error("failed to open IR output file", "err", err)
		return
	}
	defer f.Close()

	sendRc.LogConfigAndChecksum("")
	err = codec.SendIr(sendRc, f, g_senderOptions)
	if err != nil {
		slog.Error("failed to send config", "err", err)
		return
	}

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
				var settings rcconst.Settings
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

func InitScheduler(irOutputFile string, senderOptions *codec.SenderOptions) error {
	var err error

	g_irOutputFile = irOutputFile
	g_senderOptions = senderOptions

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
