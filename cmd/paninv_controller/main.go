package main

import (
	"flag"
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/utils"
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
	// var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC transmit device")
	var vRcDb = flag.String("db", db.GetDBPath(), "SQLite database")
	var vLogLevel = flag.String("log-level", "info", "log level [debug|info|warn|error]")
	var vHelp = flag.Bool("help", false, "print usage")

	var vMessage = flag.Bool("msg", false, "print message")
	var vBytes = flag.Bool("bytes", false, "print message as bytes")
	var vConfig = flag.Bool("config", false, "print decoded configuration")

	recOptions := codec.NewReceiverOptions()
	var vRaw = flag.Bool("rec-raw", recOptions.PrintRaw, "receive option: print raw pulse data")
	var vClean = flag.Bool("rec-clean", recOptions.PrintClean, "receive option: print cleaned up pulse data")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *vIrInput == "" {
		slog.Error("please set the device or file to read from")
		os.Exit(1)
	}

	slog.New(slog.NewTextHandler(os.Stdout, utils.SetLoggerOpts(*vLogLevel)))

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

	err := codec.StartIrReceiver(*vIrInput, messageHandler(options), recOptions)
	if err != nil {
		slog.Error("failed to start IR receiver", "error", err)
		os.Exit(1)
	}
}
