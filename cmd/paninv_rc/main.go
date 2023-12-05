package main

import (
	"flag"
	"fmt"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/db"
	"rpi_panasonic_inverter_rc/rclogic"
)

func main() {
	var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC transmit device or file")
	var vTrace = flag.Bool("trace", false, "print some trace output")
	var vMode2 = flag.Bool("mode2", false, "output to file in mode2 format for ending with it-ctl")
	var vIrDb = flag.String("db", "paninv.db", "SQLite database")
	var vDevice = flag.Bool("dev", true, "writing to a LIRC device")
	var vHelp = flag.Bool("help", false, "print usage")

	var vPower = flag.String("power", "", "power [on|off]")
	var vMode = flag.String("mode", "", "mode [auto|heat|cool|dry]")
	var vPowerful = flag.String("powerful", "", "powerful [on|off]")
	var vQuiet = flag.String("quiet", "", "quiet [on|off]")
	var vTemp = flag.Int("temp", 0, "temperature (uses saved mode temp when unset)")
	var vFan = flag.String("fan", "", "fan speed [auto|lowest|low|middle|high|highest]")
	var vVert = flag.String("vert", "", "vent vertical position [auto|lowest|low|middle|high|highest]")
	var vHoriz = flag.String("horiz", "", "vent horizontal position [auto|farleft|left|middle|right|farright]")
	var vTimerOnEnabled = flag.String("timeron", "", "time on [on|off]")
	var vTimerOffEnabled = flag.String("timeroff", "", "time off [on|off]")
	var vTimeOn = flag.String("ontime", "", "timer on time, e.g. 09:00")
	var vTimeOff = flag.String("offtime", "", "timer off time, e.g. 21:00")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *vIrOutput == "" {
		fmt.Printf("please set the device or file to write to")
		os.Exit(1)
	}

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

	// Create a new configuration by making a copy of the current configuration. The copy contains
	// everything except the time fields, which are unset by default. The new configuration is then
	// modified according to command line arguments.
	dbc, err := db.CurrentConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ic := dbc.MakeCopy()

	if *vTrace {
		fmt.Println("config from db")
		codec.PrintConfig(dbc)
	}

	rclogic.SetPower(*vPower, ic)
	rclogic.SetMode(*vMode, ic)
	rclogic.SetPowerful(*vPowerful, ic)
	rclogic.SetQuiet(*vQuiet, ic)
	rclogic.SetTemperature(*vTemp, ic)
	rclogic.SetFanSpeed(*vFan, ic)
	rclogic.SetVerticalPosition(*vVert, ic)
	rclogic.SetHorizontalPosition(*vHoriz, ic)

	// if timers are changed in any way, time fields are initialized
	rclogic.SetTimerOnEnabled(*vTimerOnEnabled, ic, dbc)
	rclogic.SetTimerOffEnabled(*vTimerOffEnabled, ic, dbc)
	rclogic.SetTimerOn(*vTimeOn, ic, dbc)
	rclogic.SetTimerOff(*vTimeOff, ic, dbc)

	if *vTrace {
		fmt.Println("config to send")
		codec.PrintConfig(ic)
	}

	// convert the configuration to raw LIRC bytes and then send
	senderOptions := &codec.SenderOptions{
		Mode2:  *vMode2,
		Trace:  *vTrace,
		Device: *vDevice,
	}
	err = codec.SendIr(ic, f, senderOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = db.SaveConfig(ic, dbc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *vTrace {
		fmt.Printf("saved config to %s\n", *vIrDb)
	}
}
