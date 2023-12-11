package main

import (
	"flag"
	"fmt"
	"os"
	"rpi_panasonic_inverter_rc/codec"
)

type Options struct {
	Byte    bool
	Diff    bool
	Param   bool
	Verbose bool
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
	c := codec.NewIrConfig(msg)

	var checksum string
	switch msg.Frame2.VerifyChecksum() {
	case true:
		checksum = "verified"
	case false:
		checksum = "mismatch"
	}

	codec.PrintConfigAndChecksum(c, checksum)
}

func messageHandler(options *Options) func(*codec.Message) {
	prevS := ""
	return func(msg *codec.Message) {
		curS, _ := msg.Frame2.ToVerboseString()
		if options.Verbose {
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
			printParameters(msg)
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
	var vVerbose = flag.Bool("verbose", recOptions.Verbose, "print verbose output")
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
	recOptions.Verbose = *vVerbose

	options := &Options{
		Byte:    *vByte,
		Diff:    *vDiff,
		Param:   *vParam,
		Verbose: *vVerbose,
	}

	err := codec.StartReceiver(*vIrInput, messageHandler(options), recOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
