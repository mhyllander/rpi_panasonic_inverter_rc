package main

import (
	"flag"
	"fmt"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/utils"

	"golang.org/x/sys/unix"
)

func main() {
	var err error
	var vRcDb = flag.String("db", db.GetDBPath(), "SQLite database")
	var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC output device or file")
	var vShow = flag.Bool("show", false, "show the current configuration")
	var vLogLevel = flag.String("log-level", "warn", "log level [debug|info|warn|error]")
	var vVerbose = flag.Bool("verbose", false, "print verbose output")
	var vHelp = flag.Bool("help", false, "print usage")
	var vPriority = flag.Int("prio", -10, "The priority, or niceness, of the process (-20..19)")

	var vPower = flag.String("power", "", "power [on|off]")
	var vMode = flag.String("mode", "", "mode [auto|heat|cool|dry]")
	var vPowerful = flag.String("powerful", "", "powerful [on|off]")
	var vQuiet = flag.String("quiet", "", "quiet [on|off]")
	var vTemp = flag.Int("temp", 0, "temperature (set per mode)")
	var vFan = flag.String("fan", "", "fan speed (set per mode, overridden if powerful or quiet is enabled) [auto|lowest|low|middle|high|highest]")
	var vVert = flag.String("vent.vert", "", "vent vertical position [auto|lowest|low|middle|high|highest]")
	var vHoriz = flag.String("vent.horiz", "", "vent horizontal position [auto|farleft|left|middle|right|farright]")
	var vTimerOn = flag.String("timer_on", "", "timer_on [on|off]")
	var vTimerOnTime = flag.String("timer_on.time", "", "timer_on time, e.g. 09:00")
	var vTimerOff = flag.String("timer_off", "", "timer_off [on|off]")
	var vTimerOffTime = flag.String("timer_off.time", "", "timer_off time, e.g. 21:00")

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

	if *vRcDb == "" {
		fmt.Printf("please set the db name")
		os.Exit(1)
	}
	if *vIrOutput == "" {
		fmt.Printf("please set the device or file to write to")
		os.Exit(1)
	}

	err = unix.Setpriority(unix.PRIO_PROCESS, 0, *vPriority)
	if err != nil {
		fmt.Println(err)
	}

	utils.InitLogger(*vLogLevel)

	// open and initialize database
	db.Initialize(*vRcDb)
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
	dbRc, err := db.CurrentConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *vShow {
		dbRc.PrintConfigAndChecksum("")
		os.Exit(0)
	}

	if *vVerbose {
		fmt.Println("config from db")
		dbRc.PrintConfigAndChecksum("")
	}

	// Create a new configuration by making a copy of the current configuration. The copy contains
	// everything except the time fields, which are unset by default. The new configuration is then
	// modified according to command line arguments.
	sendRc := dbRc.CopyForSending()
	utils.SetMode(*vMode, sendRc)
	utils.SetPowerful(*vPowerful, sendRc)
	utils.SetQuiet(*vQuiet, sendRc)
	utils.SetTemperature(*vTemp, sendRc)
	utils.SetFanSpeed(*vFan, sendRc)
	utils.SetVentVerticalPosition(*vVert, sendRc)
	utils.SetVentHorizontalPosition(*vHoriz, sendRc)

	// if timers are changed in any way, time fields are initialized
	utils.SetTimerOn(*vTimerOn, sendRc, dbRc)
	utils.SetTimerOnTime(*vTimerOnTime, sendRc, dbRc)
	utils.SetTimerOff(*vTimerOff, sendRc, dbRc)
	utils.SetTimerOffTime(*vTimerOffTime, sendRc, dbRc)

	// set power last, adjusting for any current timers
	utils.SetPower(*vPower, sendRc, dbRc)

	if *vVerbose {
		fmt.Println("config to send")
		sendRc.PrintConfigAndChecksum("")
	}

	senderOptions.Mode2 = *vMode2
	senderOptions.Device = *vDevice
	senderOptions.Transmissions = *vTransmissions
	senderOptions.Interval_ms = *vInterval

	err = codec.SendIr(sendRc, f, senderOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *vVerbose {
		fmt.Printf("saved config to %s\n", *vRcDb)
	}
}
