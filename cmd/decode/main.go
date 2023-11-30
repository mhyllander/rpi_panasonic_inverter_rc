package main

import (
	"flag"
	"fmt"
	"os"
	"panasonic_irc/codec"
)

func printMessageDiff(prevS, curS string) {
	diffS := "   "
	cmp := false
	for i := 0; i < len(curS); i++ {
		if cmp && prevS[i] != curS[i] {
			diffS += "^"
		} else {
			diffS += " "
		}
		switch curS[i] {
		case '[':
			cmp = true
		case ']':
			cmp = false
		}
	}
	fmt.Println(diffS)
}

func printParams(f codec.Frame) {
	power := f.ValueFromOneByte(codec.PANASONIC_POWER_BYTE, codec.PANASONIC_POWER_MASK)
	temp := f.ValueFromOneByte(codec.PANASONIC_TEMP_BYTE, codec.PANASONIC_TEMP_MASK)
	mode := f.ValueFromOneByte(codec.PANASONIC_MODE_BYTE, codec.PANASONIC_MODE_MASK)
	fan := f.ValueFromOneByte(codec.PANASONIC_FAN_BYTE, codec.PANASONIC_FAN_MASK)
	quiet := f.ValueFromOneByte(codec.PANASONIC_QUIET_BYTE, codec.PANASONIC_QUIET_MASK)
	powerful := f.ValueFromOneByte(codec.PANASONIC_POWERFUL_BYTE, codec.PANASONIC_POWERFUL_MASK)
	hpos := f.ValueFromOneByte(codec.PANASONIC_VENT_HPOS_BYTE, codec.PANASONIC_VENT_HPOS_MASK)
	vpos := f.ValueFromOneByte(codec.PANASONIC_VENT_VPOS_BYTE, codec.PANASONIC_VENT_VPOS_MASK)

	fmt.Printf("power=%d mode=%d temp=%d fan=%d quiet=%d powerful=%d hpos=%d vpos=%d\n",
		power, mode, temp, fan, quiet, powerful, hpos, vpos)

	clock := f.ValueFromTwoBytes(codec.PANASONIC_CLOCK_BYTE1, codec.PANASONIC_CLOCK_MASK1, codec.PANASONIC_CLOCK_BYTE2, codec.PANASONIC_CLOCK_MASK2)
	clock_set := clock != codec.PANASONIC_TIME_UNSET

	timer_on_enabled := f.ValueFromOneByte(codec.PANASONIC_TIMER_ON_ENABLED_BYTE, codec.PANASONIC_TIMER_ON_ENABLED_MASK)
	timer_on_time := f.ValueFromTwoBytes(codec.PANASONIC_TIMER_ON_TIME_BYTE1, codec.PANASONIC_TIMER_ON_TIME_MASK1, codec.PANASONIC_TIMER_ON_TIME_BYTE2, codec.PANASONIC_TIMER_ON_TIME_MASK2)
	timer_on_time_set := timer_on_time != codec.PANASONIC_TIME_UNSET

	timer_off_enabled := f.ValueFromOneByte(codec.PANASONIC_TIMER_OFF_ENABLED_BYTE, codec.PANASONIC_TIMER_OFF_ENABLED_MASK)
	timer_off_time := f.ValueFromTwoBytes(codec.PANASONIC_TIMER_OFF_TIME_BYTE1, codec.PANASONIC_TIMER_OFF_TIME_MASK1, codec.PANASONIC_TIMER_OFF_TIME_BYTE2, codec.PANASONIC_TIMER_OFF_TIME_MASK2)
	timer_off_time_set := timer_off_time != codec.PANASONIC_TIME_UNSET

	fmt.Printf(
		"Clock set=%t time=%02d:%02d, Timer_On enabled=%d set=%t time=%02d:%02d, Timer_Off enabled=%d set=%t time=%02d:%02d\n",
		clock_set, clock/60, clock%60,
		timer_on_enabled, timer_on_time_set, timer_on_time/60, timer_on_time%60,
		timer_off_enabled, timer_off_time_set, timer_off_time/60, timer_off_time%60,
	)
}

func messageProcessor(options *codec.ReaderOptions) func([]codec.Frame) {
	prevS := ""
	return func(msg []codec.Frame) {
		curS := msg[1].ToTraceString()
		if options.Trace {
			fmt.Printf("Message, frames=%d\n", len(msg))
			fmt.Printf("%d: %s\n", 1, msg[0].ToTraceString())
		}
		if options.Trace || options.Diff {
			fmt.Printf("%d: %s\n", 2, curS)
		}
		if options.Diff && prevS != "" {
			// compare current frames with previous
			printMessageDiff(prevS, curS)
		}
		if options.Param {
			printParams(msg[1])
		}
		prevS = curS
	}
}

func main() {
	var vFile = flag.String("file", "/dev/lirc-rx", "file to parse")
	var vSocket = flag.Bool("sock", false, "read from socket")
	var vRaw = flag.Bool("raw", false, "print raw pulse data")
	var vClean = flag.Bool("clean", false, "print cleaned up pulse data")
	var vTrace = flag.Bool("trace", false, "print message trace")
	var vDiff = flag.Bool("diff", false, "show difference from previous")
	var vParam = flag.Bool("param", false, "show decoded params")
	var vHelp = flag.Bool("help", false, "print usage")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *vFile == "" {
		fmt.Printf("please set the file to read from")
		os.Exit(1)
	}

	options := codec.ReaderOptions{
		Socket: *vSocket,
		Raw:    *vRaw,
		Clean:  *vClean,
		Trace:  *vTrace,
		Diff:   *vDiff,
		Param:  *vParam,
	}

	err := codec.StartReader(*vFile, messageProcessor(&options), &options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
