package main

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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

		err := db.UpdateConfig(
			func(current *codec.RcConfig) (*codec.RcConfig, error) {
				return c, nil
			},
			nil,
		)
		if err != nil {
			slog.Error("failed to save the new config", "error", err)
			return
		}

		if err := sched.RestartTimerJobs(); err != nil {
			slog.Error("failed to restart timer jobs", "err", err)
		}
	}
}

type jobSet struct {
	Active   bool      `json:"active"`
	CronJobs []cronJob `json:"cronjobs"`
}

type cronJob struct {
	Name     string             `json:"name,omitempty"`
	Schedule string             `json:"schedule"`
	Settings codecbase.Settings `json:"settings"`
}

type jobSets map[string]jobSet

func importJobs(jobfile string) error {
	f, err := os.Open(jobfile)
	if err != nil {
		slog.Error("failed to open jobs file", "err", err)
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		slog.Error("failed to read jobs file", "err", err)
		return err
	}

	var jobsets jobSets
	err = json.Unmarshal(data, &jobsets)
	if err != nil {
		slog.Error("failed to unmarshal jobsets", "err", err)
		return err
	}

	var definitions []db.JobSetDefinition
	for name, js := range jobsets {
		definition := db.JobSetDefinition{Name: name, Active: js.Active}
		for _, j := range js.CronJobs {
			definition.CronJobs = append(definition.CronJobs, db.CronJobDefinition{Name: j.Name, Schedule: j.Schedule, Settings: j.Settings})
		}
		definitions = append(definitions, definition)
	}

	if err := db.ReplaceJobSets(definitions); err != nil {
		slog.Error("failed to load jobs, existing jobs left unchanged", "err", err)
		return err
	}
	return nil
}

// exportJobs writes all jobsets as JSON, in the format read by importJobs. An empty file name or "-" means stdout.
func exportJobs(jobfile string) error {
	definitions, err := db.GetJobSetDefinitions()
	if err != nil {
		slog.Error("failed to get jobsets", "err", err)
		return err
	}

	exported := make(jobSets, len(definitions))
	for _, definition := range definitions {
		js := jobSet{Active: definition.Active, CronJobs: make([]cronJob, 0, len(definition.CronJobs))}
		for _, j := range definition.CronJobs {
			js.CronJobs = append(js.CronJobs, cronJob{Name: j.Name, Schedule: j.Schedule, Settings: j.Settings})
		}
		exported[definition.Name] = js
	}

	data, err := json.Marshal(exported, jsontext.Multiline(true))
	if err != nil {
		slog.Error("failed to marshal jobsets", "err", err)
		return err
	}
	data = append(data, '\n')

	if jobfile == "" || jobfile == "-" {
		_, err = os.Stdout.Write(data)
	} else {
		err = os.WriteFile(jobfile, data, 0o644)
	}
	if err != nil {
		slog.Error("failed to write jobs", "err", err)
	}
	return err
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

	var vImportJobs = flag.String("import-jobs", "", "replace all jobsets with the jobsets in a JSON file")
	var vExportJobs = flag.String("export-jobs", "", "write all jobsets as JSON to a file, or to stdout if the file is -")

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
	if err := db.Initialize(*vRcDb); err != nil {
		slog.Error("failed to initialize database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if *vImportJobs != "" && *vExportJobs != "" {
		slog.Error("-import-jobs and -export-jobs can't be used together")
		os.Exit(1)
	}
	if *vImportJobs != "" {
		if err := importJobs(*vImportJobs); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	if *vExportJobs != "" {
		if err := exportJobs(*vExportJobs); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	// start the IR receiver
	irReceiver := codec.NewIrReceiver(*vIrInput, messageHandler(&options), recOptions)
	go func() {
		// this call blocks
		err := irReceiver.Run()
		if err != nil {
			slog.Error("failed to start IR receiver", "err", err)
		}
	}()
	defer irReceiver.Quit()

	irSender := codec.StartIrSender(*vIrOutput, senderOptions, irReceiver)
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
