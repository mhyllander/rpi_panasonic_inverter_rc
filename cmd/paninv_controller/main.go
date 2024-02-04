package main

import (
	"flag"
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/sched"
	"rpi_panasonic_inverter_rc/utils"
	"time"
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

		c.LogConfigAndChecksum(checksum)
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
		slog.Debug("saved config to db")
	}
}

func main() {
	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC receive device")
	var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC transmit device")
	var vRcDb = flag.String("db", db.GetDBPath(), "SQLite database")
	var vLogLevel = flag.String("log-level", "info", "log level [debug|info|warn|error]")
	var vHelp = flag.Bool("help", false, "print usage")

	var vMessage = flag.Bool("msg", false, "print message")
	var vBytes = flag.Bool("bytes", false, "print message as bytes")
	var vConfig = flag.Bool("config", false, "print decoded configuration")

	recOptions := codec.NewReceiverOptions()
	var vRaw = flag.Bool("rec-raw", recOptions.PrintRaw, "receive option: print raw pulse data")
	var vClean = flag.Bool("rec-clean", recOptions.PrintClean, "receive option: print cleaned up pulse data")

	senderOptions := codec.NewSenderOptions()
	var vMode2 = flag.Bool("send-mode2", senderOptions.Mode2, "send option: output in mode2 format (when writing to file for sending with ir-ctl)")
	var vTransmissions = flag.Int("send-tx", senderOptions.Transmissions, "send option: number of times to send the message")
	var vInterval = flag.Int("send-int", senderOptions.Interval_ms, "send option: number of milliseconds between transmissions")
	var vDevice = flag.Bool("send-dev", senderOptions.Device, "send option: writing to a LIRC device")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	utils.InitLogger(*vLogLevel)

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

	recOptions.Device = true
	recOptions.PrintRaw = *vRaw
	recOptions.PrintClean = *vClean

	options := &Options{
		PrintBytes:   *vBytes,
		PrintConfig:  *vConfig,
		PrintMessage: *vMessage,
	}

	senderOptions.Mode2 = *vMode2
	senderOptions.Device = *vDevice
	senderOptions.Transmissions = *vTransmissions
	senderOptions.Interval_ms = *vInterval

	// start the IR receiver
	go func() {
		// this call blocks
		err := codec.StartIrReceiver(*vIrInput, messageHandler(options), recOptions)
		if err != nil {
			slog.Error("failed to start IR receiver", "err", err)
		}
	}()

	// start gocron
	err := sched.InitScheduler(*vIrOutput, senderOptions)
	if err != nil {
		slog.Error("failed to start scheduler", "error", err)
	}
	defer sched.Stop()

	// wait here
	for {
		time.Sleep(time.Second)
	}
}
