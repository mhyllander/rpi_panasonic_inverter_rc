package sched

import (
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
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
	codec.SuspendReceiver()
	defer codec.ResumeReceiver()

	slog.Info("initializing: sending current config")

	dbRc, err := db.CurrentConfigForSending()
	if err != nil {
		slog.Error("failed to get current config", "err", err)
		return
	}
	f := openIrOutputFile()
	if f == nil {
		slog.Error("failed to open IR output file", "err", err)
		return
	}
	defer f.Close()

	err = codec.SendIr(dbRc, f, g_senderOptions)
	if err != nil {
		slog.Error("failed to send current config", "err", err)
		return
	}
}

func InitScheduler(irOutputFile string, senderOptions *codec.SenderOptions) error {
	g_irOutputFile = irOutputFile
	g_senderOptions = senderOptions

	// create a scheduler
	scheduler, err := gocron.NewScheduler(
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
			gocron.OneTimeJobStartDateTime(time.Now().Add(time.Minute)),
		),
		gocron.NewTask(
			SendCurrentConfig,
		),
	)
	if err != nil {
		return err
	}
	slog.Debug("scheduled initial job", "job_id", j.ID())

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
