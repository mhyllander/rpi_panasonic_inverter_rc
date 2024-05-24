package main

import (
	"flag"
	"fmt"
	"os"

	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/logs"
)

type Options struct {
	PrintBytes   bool
	PrintDiff    bool
	PrintConfig  bool
	PrintMessage bool
}

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

func printParameters(msg *codec.Message) {
	c := codec.RcConfigFromFrame(msg)

	var checksum string
	switch msg.Frame2.VerifyChecksum() {
	case true:
		checksum = "verified"
	case false:
		checksum = "mismatch"
	}

	c.PrintConfigAndChecksum(checksum)
}

func messageHandler(options *Options) func(*codec.Message) {
	prevS := ""
	return func(msg *codec.Message) {
		curS, _ := msg.Frame2.ToVerboseString()
		if options.PrintMessage {
			msg.PrintMessage()
		}
		if options.PrintDiff && prevS != "" {
			// compare current frames with previous
			printMessageDiff(prevS, curS)
		}
		if options.PrintBytes {
			msg.PrintByteRepresentation()
		}
		if options.PrintConfig {
			printParameters(msg)
		}
		prevS = curS
	}
}

func main() {
	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC source (file or device)")
	var vLogLevel = flag.String("log-level", "debug", "log level [debug|info|warn|error]")
	var vHelp = flag.Bool("help", false, "print usage")

	var options Options
	flag.BoolVar(&options.PrintMessage, "msg", false, "print message")
	flag.BoolVar(&options.PrintBytes, "bytes", false, "print message as bytes")
	flag.BoolVar(&options.PrintDiff, "diff", false, "print difference from previous")
	flag.BoolVar(&options.PrintConfig, "config", false, "print decoded configuration")

	recOptions := codec.NewReceiverOptions()
	flag.BoolVar(&recOptions.Device, "rec-dev", recOptions.Device, "receive option: reading from LIRC device")
	flag.BoolVar(&recOptions.PrintRaw, "rec-raw", recOptions.PrintRaw, "receive option: print raw pulse data")
	flag.BoolVar(&recOptions.PrintClean, "rec-clean", recOptions.PrintClean, "receive option: print cleaned up pulse data")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *vIrInput == "" {
		fmt.Printf("please set the device or file to read from")
		os.Exit(1)
	}

	logs.InitLogger(*vLogLevel)

	// this call blocks
	err := codec.RunIrReceiver(*vIrInput, messageHandler(&options), recOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
