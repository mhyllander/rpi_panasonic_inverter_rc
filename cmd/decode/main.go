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
	power := f.GetValue(codec.PANASONIC_POWER_BIT0, codec.PANASONIC_POWER_BITS)
	temp := f.GetValue(codec.PANASONIC_TEMP_BIT0, codec.PANASONIC_TEMP_BITS)
	mode := f.GetValue(codec.PANASONIC_MODE_BIT0, codec.PANASONIC_MODE_BITS)
	fan := f.GetValue(codec.PANASONIC_FAN_BIT0, codec.PANASONIC_FAN_BITS)
	quiet := f.GetValue(codec.PANASONIC_QUIET_BIT0, codec.PANASONIC_QUIET_BITS)
	powerful := f.GetValue(codec.PANASONIC_POWERFUL_BIT0, codec.PANASONIC_POWERFUL_BITS)
	hpos := f.GetValue(codec.PANASONIC_VENT_HPOS_BIT0, codec.PANASONIC_VENT_HPOS_BITS)
	vpos := f.GetValue(codec.PANASONIC_VENT_VPOS_BIT0, codec.PANASONIC_VENT_VPOS_BITS)

	fmt.Printf("power=%d mode=%d temp=%d fan=%d quiet=%d powerful=%d hpos=%d vpos=%d\n",
		power, mode, temp, fan, quiet, powerful, hpos, vpos)

	clock := f.GetValue(codec.PANASONIC_CLOCK_BIT0, codec.PANASONIC_CLOCK_BITS)

	timer_on_enabled := f.GetValue(codec.PANASONIC_TIMER_ON_ENABLED_BIT0, codec.PANASONIC_TIMER_ON_ENABLED_BITS)
	timer_on_time := f.GetValue(codec.PANASONIC_TIMER_ON_TIME_BIT0, codec.PANASONIC_TIMER_ON_TIME_BITS)

	timer_off_enabled := f.GetValue(codec.PANASONIC_TIMER_OFF_ENABLED_BIT0, codec.PANASONIC_TIMER_OFF_ENABLED_BITS)
	timer_off_time := f.GetValue(codec.PANASONIC_TIMER_OFF_TIME_BIT0, codec.PANASONIC_TIMER_OFF_TIME_BITS)

	fmt.Printf(
		"Clock time=%02d:%02d, Timer_On enabled=%d time=%02d:%02d, Timer_Off enabled=%d time=%02d:%02d\n",
		clock/60, clock%60,
		timer_on_enabled, timer_on_time/60, timer_on_time%60,
		timer_off_enabled, timer_off_time/60, timer_off_time%60,
	)
}

func messageProcessor(options *codec.ReaderOptions) func(*codec.Message) {
	prevS := ""
	return func(msg *codec.Message) {
		curS := msg.Frame2.ToTraceString()
		if options.Trace {
			fmt.Printf("Message\n")
			fmt.Printf("%d: %s\n", 1, msg.Frame1.ToTraceString())
		}
		if options.Trace || options.Diff {
			fmt.Printf("%d: %s\n", 2, curS)
		}
		if options.Diff && prevS != "" {
			// compare current frames with previous
			printMessageDiff(prevS, curS)
		}
		if options.Trace {
			fmt.Println("Byte representation of BitSet:")
			fmt.Printf("  %d: %s\n", 1, msg.Frame1.ToByteString())
			fmt.Printf("  %d: %s\n", 2, msg.Frame2.ToByteString())
		}
		if options.Param {
			printParams(msg.Frame2)
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
