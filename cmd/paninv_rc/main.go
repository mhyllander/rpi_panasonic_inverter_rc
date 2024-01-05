package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/utils"
)

func main() {
	var vIrDb = flag.String("db", db.GetDBPath(), "SQLite database")
	var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC output device or file")
	var vShow = flag.Bool("show", false, "show the current configuration")
	var vLogLevel = flag.String("log-level", "warn", "log level [debug|info|warn|error]")
	var vVerbose = flag.Bool("verbose", false, "print verbose output")
	var vHelp = flag.Bool("help", false, "print usage")

	var vPower = flag.String("power", "", "power [on|off]")
	var vMode = flag.String("mode", "", "mode [auto|heat|cool|dry]")
	var vPowerful = flag.String("powerful", "", "powerful [on|off]")
	var vQuiet = flag.String("quiet", "", "quiet [on|off]")
	var vTemp = flag.Int("temp", 0, "temperature (set per mode)")
	var vFan = flag.String("fan", "", "fan speed (set per mode, overridden if powerful or quiet is enabled) [auto|lowest|low|middle|high|highest]")
	var vVert = flag.String("vert", "", "vent vertical position [auto|lowest|low|middle|high|highest]")
	var vHoriz = flag.String("horiz", "", "vent horizontal position [auto|farleft|left|middle|right|farright]")
	var vTimerOnEnabled = flag.String("ton", "", "timer on [on|off]")
	var vTimerOffEnabled = flag.String("toff", "", "timer off [on|off]")
	var vTimeOn = flag.String("ton-time", "", "timer on time, e.g. 09:00")
	var vTimeOff = flag.String("toff-time", "", "timer off time, e.g. 21:00")

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

	if *vIrDb == "" {
		fmt.Printf("please set the db name")
		os.Exit(1)
	}
	if *vIrOutput == "" {
		fmt.Printf("please set the device or file to write to")
		os.Exit(1)
	}

	slog.New(slog.NewTextHandler(os.Stdout, utils.SetLoggerOpts(*vLogLevel)))

	// open and initialize database
	db.Initialize(*vIrDb)
	defer db.Close()

	// open file or device for sending IR
	flags := os.O_RDWR
	if !*vDevice {
		flags = flags | os.O_CREATE
	}
	f, err := os.OpenFile(*vIrOutput, flags, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	// get current configuration
	dbIc, err := db.CurrentConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *vShow {
		codec.PrintConfigAndChecksum(dbIc, "")
		os.Exit(0)
	}

	if *vVerbose {
		fmt.Println("config from db")
		codec.PrintConfigAndChecksum(dbIc, "")
	}

	// Create a new configuration by making a copy of the current configuration. The copy contains
	// everything except the time fields, which are unset by default. The new configuration is then
	// modified according to command line arguments.
	sendIc := dbIc.CopyForSending()
	utils.SetMode(*vMode, sendIc)
	utils.SetPowerful(*vPowerful, sendIc)
	utils.SetQuiet(*vQuiet, sendIc)
	utils.SetTemperature(*vTemp, sendIc)
	utils.SetFanSpeed(*vFan, sendIc)
	utils.SetVentVerticalPosition(*vVert, sendIc)
	utils.SetVentHorizontalPosition(*vHoriz, sendIc)

	// set power, adjusting for any current timers
	utils.SetPower(*vPower, sendIc, dbIc)

	// if timers are changed in any way, time fields are initialized
	utils.SetTimerOnEnabled(*vTimerOnEnabled, sendIc, dbIc)
	utils.SetTimerOffEnabled(*vTimerOffEnabled, sendIc, dbIc)
	utils.SetTimerOn(*vTimeOn, sendIc, dbIc)
	utils.SetTimerOff(*vTimeOff, sendIc, dbIc)

	if *vVerbose {
		fmt.Println("config to send")
		codec.PrintConfigAndChecksum(sendIc, "")
	}

	senderOptions.Mode2 = *vMode2
	senderOptions.Device = *vDevice
	senderOptions.Transmissions = *vTransmissions
	senderOptions.Interval_ms = *vInterval

	err = codec.SendIr(sendIc, f, senderOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = db.SaveConfig(sendIc, dbIc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *vVerbose {
		fmt.Printf("saved config to %s\n", *vIrDb)
	}
}
