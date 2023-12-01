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

func printParams(msg *codec.Message) {
	c := codec.NewIrConfig(msg)

	fmt.Printf("power=%d mode=%d powerful=%d quiet=%d temp=%d fan=%d vpos=%d hpos=%d\n",
		c.Power, c.Mode, c.Powerful, c.Quiet, c.Temperature, c.FanSpeed, c.VentVertical, c.VentHorizontal)

	ctime := codec.Time(msg.Frame2.GetValue(codec.PANASONIC_CLOCK_BIT0, codec.PANASONIC_CLOCK_BITS))

	fmt.Printf(
		"Timer_On enabled=%d time=%s, Timer_Off enabled=%d time=%s, Clock time=%s, \n",
		c.TimerOnEnabled, c.TimerOn, c.TimerOffEnabled, c.TimerOff, ctime)
}

func messageProcessor(options *codec.ReaderOptions) func(*codec.Message) {
	prevS := ""
	return func(msg *codec.Message) {
		curS, posS := msg.Frame2.ToTraceString()
		if options.Trace {
			fmt.Printf("Message as bit stream (first and least significant bit to the right)\n")
			t, p := msg.Frame1.ToTraceString()
			fmt.Printf("   %s\n%d: %s\n", p, 1, t)
		}
		if options.Trace || options.Diff {
			fmt.Printf("   %s\n%d: %s\n", posS, 2, curS)
		}
		if options.Diff && prevS != "" {
			// compare current frames with previous
			printMessageDiff(prevS, curS)
		}
		if options.Byte {
			fmt.Println("Byte representation:")
			fmt.Printf("  %d: %s\n", 1, msg.Frame1.ToByteString())
			fmt.Printf("  %d: %s\n", 2, msg.Frame2.ToByteString())
		}
		if options.Param {
			printParams(msg)
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
	var vByte = flag.Bool("byte", false, "print message as bytes")
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
		Byte:   *vByte,
		Diff:   *vDiff,
		Param:  *vParam,
	}

	err := codec.StartReader(*vFile, messageProcessor(&options), &options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
