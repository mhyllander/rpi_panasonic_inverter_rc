package main

import (
	"flag"
	"fmt"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/common"
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

	var settings common.Settings
	flag.StringVar(&settings.Power, "power", "", "power [on|off]")
	flag.StringVar(&settings.Mode, "mode", "", "mode [auto|heat|cool|dry]")
	flag.StringVar(&settings.Powerful, "powerful", "", "powerful [on|off]")
	flag.StringVar(&settings.Quiet, "quiet", "", "quiet [on|off]")
	flag.StringVar(&settings.Temperature, "temp", "", "temperature (set per mode)")
	flag.StringVar(&settings.FanSpeed, "fan", "", "fan speed (set per mode, overridden if powerful or quiet is enabled) [auto|lowest|low|middle|high|highest]")
	flag.StringVar(&settings.VentVertical, "vert", "", "vent vertical position [auto|lowest|low|middle|high|highest]")
	flag.StringVar(&settings.VentHorizontal, "horiz", "", "vent horizontal position [auto|farleft|left|middle|right|farright]")
	flag.StringVar(&settings.TimerOn, "ton", "", "timer_on [on|off]")
	flag.StringVar(&settings.TimerOnTime, "tont", "", "timer_on time, e.g. 09:00")
	flag.StringVar(&settings.TimerOff, "toff", "", "timer_off [on|off]")
	flag.StringVar(&settings.TimerOffTime, "tofft", "", "timer_off time, e.g. 21:00")

	senderOptions := codec.NewSenderOptions()
	flag.BoolVar(&senderOptions.Mode2, "send-mode2", senderOptions.Mode2, "send option: output in mode2 format (when writing to file for sending with ir-ctl)")
	flag.IntVar(&senderOptions.Transmissions, "send-tx", senderOptions.Transmissions, "send option: number of times to send the message")
	flag.IntVar(&senderOptions.Interval_ms, "send-int", senderOptions.Interval_ms, "send option: number of milliseconds between transmissions")
	flag.BoolVar(&senderOptions.Device, "send-dev", senderOptions.Device, "send option: writing to a LIRC device")

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
	sendRc := utils.ComposeSendConfig(&settings, dbRc)

	if *vVerbose {
		fmt.Println("config to send")
		sendRc.PrintConfigAndChecksum("")
	}

	irSender := codec.StartIrSender(*vIrOutput, senderOptions)
	irSender.SendConfig(sendRc)
	irSender.Stop()

	err = db.SaveConfig(sendRc, dbRc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *vVerbose {
		fmt.Printf("saved config to %s\n", *vRcDb)
	}
}
