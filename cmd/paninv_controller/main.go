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
			codec.PrintMessage(msg)
		}
		if options.PrintBytes {
			codec.PrintByteRepresentation(msg)
		}

		c := codec.NewIrConfig(msg)

		if options.PrintConfig {
			codec.PrintConfigAndChecksum(c, checksum)
		}

		// get current configuration
		dbIc, err := db.CurrentConfig()
		if err != nil {
			slog.Error("filed to get current config", "error", err)
			return
		}

		err = db.SaveConfig(c, dbIc)
		if err != nil {
			slog.Error("failed to save the new config", "error", err)
			return
		}
		if options.PrintMessage {
			slog.Debug("saved config to db")
		}
	}
}

func main() {
	recOptions := codec.NewReceiverOptions()

	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC receive device ")
	// var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC transmit device")
	var vIrDb = flag.String("db", "paninv.db", "SQLite database")
	var vLogLevel = flag.String("log-level", "info", "log level [debug|info|warn|error]")
	var vHelp = flag.Bool("help", false, "print usage")

	var vDevice = flag.Bool("dev", recOptions.Device, "receive option: reading from LIRC device")
	var vRaw = flag.Bool("raw", recOptions.PrintRaw, "receive option: print raw pulse data")
	var vClean = flag.Bool("clean", recOptions.PrintClean, "receive option: print cleaned up pulse data")

	var vMessage = flag.Bool("msg", false, "print message")
	var vBytes = flag.Bool("bytes", false, "print message as bytes")
	var vConfig = flag.Bool("config", false, "print decoded configuration")

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
	db.Initialize(*vIrDb)
	defer db.Close()

	recOptions.Device = *vDevice
	recOptions.PrintRaw = *vRaw
	recOptions.PrintClean = *vClean

	options := &Options{
		PrintBytes:   *vBytes,
		PrintConfig:  *vConfig,
		PrintMessage: *vMessage,
	}

	err := codec.StartReceiver(*vIrInput, messageHandler(options), recOptions)
	if err != nil {
		slog.Error("failed to start IR receiver", "error", err)
		os.Exit(1)
	}
}
