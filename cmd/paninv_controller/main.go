package main

import (
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"os"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/codecbase"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/logs"
	"rpi_panasonic_inverter_rc/sched"
	"rpi_panasonic_inverter_rc/server"
)

type Options struct {
	PrintBytes   bool
	PrintConfig  bool
	PrintMessage bool
}

func messageHandler(options *Options) func(*codec.Message) {
	return func(msg *codec.Message) {
		var checksum string
		switch msg.Frame2.VerifyChecksum() {
		case true:
			checksum = "verified"
		case false:
			checksum = "mismatch"
		}

		if options.PrintMessage {
			msg.PrintMessage()
		}
		if options.PrintBytes {
			msg.PrintByteRepresentation()
		}

		c := codec.RcConfigFromFrame(msg)

		c.LogConfigAndChecksum("received config", checksum)
		if options.PrintConfig {
			c.PrintConfigAndChecksum(checksum)
		}

		if checksum != "verified" {
			slog.Warn("checksum mismatch, discarding")
			return
		}

		// get current configuration
		dbRc, err := db.CurrentConfig()
		if err != nil {
			slog.Error("failed to get current config", "error", err)
			return
		}

		err = db.SaveConfig(c, dbRc)
		if err != nil {
			slog.Error("failed to save the new config", "error", err)
			return
		}

		sched.RestartTimerJobs()
	}
}

type jobSet struct {
	Active   bool      `json:"active"`
	CronJobs []cronJob `json:"cronjobs"`
}

type cronJob struct {
	Schedule string             `json:"schedule"`
	Settings codecbase.Settings `json:"settings"`
}

type jobSets map[string]jobSet

func loadCronJobs(jobfile string) {
	f, err := os.Open(jobfile)
	if err != nil {
		slog.Error("failed to open jobs file")
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		slog.Error("failed to read jobs file")
		return
	}

	var jobsets jobSets
	err = json.Unmarshal(data, &jobsets)
	if err != nil {
		slog.Error("failed to unmarshal jobsets")
		return
	}

	db.DeleteAllJobSetsPermanently()
	db.DeleteAllCronJobsPermanently()
	for name, js := range jobsets {
		if err := db.SaveJobSet(name, js.Active); err != nil {
			slog.Error("failed to save jobset", "jobset", name, "err", err)
		}
		for _, j := range js.CronJobs {
			if err := db.SaveCronJob(name, j.Schedule, &j.Settings); err != nil {
				slog.Error("failed to save cronjob", "err", err)
			}
		}
	}
}

func main() {
	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC receive device")
	var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC transmit device")
	var vRcDb = flag.String("db", db.GetDBPath(), "SQLite database")
	var vLogLevel = flag.String("log-level", "info", "log level [debug|info|warn|error]")
	var vHelp = flag.Bool("help", false, "print usage")

	var options Options
	flag.BoolVar(&options.PrintMessage, "msg", false, "print message")
	flag.BoolVar(&options.PrintBytes, "bytes", false, "print message as bytes")
	flag.BoolVar(&options.PrintConfig, "config", false, "print decoded configuration")

	recOptions := codec.NewReceiverOptions()
	flag.BoolVar(&recOptions.PrintRaw, "rec-raw", recOptions.PrintRaw, "receive option: print raw pulse data")
	flag.BoolVar(&recOptions.PrintClean, "rec-clean", recOptions.PrintClean, "receive option: print cleaned up pulse data")

	senderOptions := codec.NewSenderOptions()
	flag.BoolVar(&senderOptions.Mode2, "send-mode2", senderOptions.Mode2, "send option: output in mode2 format (when writing to file for sending with ir-ctl)")
	flag.IntVar(&senderOptions.Transmissions, "send-tx", senderOptions.Transmissions, "send option: number of times to send the message")
	flag.IntVar(&senderOptions.Interval_ms, "send-int", senderOptions.Interval_ms, "send option: number of milliseconds between transmissions")
	flag.BoolVar(&senderOptions.Device, "send-dev", senderOptions.Device, "send option: writing to a LIRC device")

	var vLoadJobs = flag.String("load-jobs", "", "load cronjobs from file")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	logs.InitLogger(*vLogLevel)

	if *vRcDb == "" {
		slog.Error("please set the db name")
		os.Exit(1)
	}
	if *vIrInput == "" {
		slog.Error("please set the device or file to read from")
		os.Exit(1)
	}
	if *vIrOutput == "" {
		slog.Error("please set the device or file to write to")
		os.Exit(1)
	}

	// open and initialize database
	db.Initialize(*vRcDb)
	defer db.Close()

	if *vLoadJobs != "" {
		loadCronJobs(*vLoadJobs)
		os.Exit(0)
	}

	// start the IR receiver
	go func() {
		// this call blocks
		err := codec.RunIrReceiver(*vIrInput, messageHandler(&options), recOptions)
		if err != nil {
			slog.Error("failed to start IR receiver", "err", err)
		}
	}()

	irSender := codec.StartIrSender(*vIrOutput, senderOptions)
	defer irSender.Stop()

	// start gocron
	err := sched.InitScheduler(irSender)
	if err != nil {
		slog.Error("failed to start scheduler", "error", err)
	}
	defer sched.Stop()

	// Start web server
	server.StartServer(*vLogLevel, irSender)
}
