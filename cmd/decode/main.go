package main

import (
	"flag"
	"fmt"
	"os"
	"rpi_panasonic_inverter_rc/codec"
)

type Options struct {
	Byte  bool
	Diff  bool
	Param bool
	Trace bool
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

func messageHandler(options *Options) func(*codec.Message) {
	prevS := ""
	return func(msg *codec.Message) {
		curS, _ := msg.Frame2.ToTraceString()
		if options.Trace {
			codec.PrintMessage(msg)
		}
		if options.Diff && prevS != "" {
			// compare current frames with previous
			printMessageDiff(prevS, curS)
		}
		if options.Byte {
			codec.PrintByteRepresentation(msg)
		}
		if options.Param {
			codec.PrintParams(msg)
		}
		prevS = curS
	}
}

func main() {
	recOptions := codec.NewReceiverOptions()

	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC data source (file or device)")
	var vHelp = flag.Bool("help", false, "print usage")

	var vDevice = flag.Bool("dev", recOptions.Device, "reading from LIRC device")
	var vRaw = flag.Bool("raw", recOptions.Raw, "print raw pulse data")
	var vClean = flag.Bool("clean", recOptions.Clean, "print cleaned up pulse data")
	var vTrace = flag.Bool("trace", recOptions.Trace, "print message trace")
	var vByte = flag.Bool("byte", false, "print message as bytes")
	var vDiff = flag.Bool("diff", false, "show difference from previous")
	var vParam = flag.Bool("param", false, "show decoded params")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *vIrInput == "" {
		fmt.Printf("please set the device or file to read from")
		os.Exit(1)
	}

	recOptions.Device = *vDevice
	recOptions.Raw = *vRaw
	recOptions.Clean = *vClean
	recOptions.Trace = *vTrace

	options := &Options{
		Byte:  *vByte,
		Diff:  *vDiff,
		Param: *vParam,
		Trace: *vTrace,
	}

	err := codec.StartReceiver(*vIrInput, messageHandler(options), recOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
