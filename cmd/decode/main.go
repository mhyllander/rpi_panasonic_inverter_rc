package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/utils"
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
		if options.PrintMessage {
			codec.PrintMessage(msg)
		}
		if options.PrintDiff && prevS != "" {
			// compare current frames with previous
			printMessageDiff(prevS, curS)
		}
		if options.PrintBytes {
			codec.PrintByteRepresentation(msg)
		}
		if options.PrintConfig {
			printParameters(msg)
		}
		prevS = curS
	}
}

func main() {
	recOptions := codec.NewReceiverOptions()

	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC data source (file or device)")
	var vLogLevel = flag.String("log-level", "debug", "log level [debug|info|warn|error]")
	var vHelp = flag.Bool("help", false, "print usage")

	var vDevice = flag.Bool("dev", recOptions.Device, "receive option: reading from LIRC device")
	var vRaw = flag.Bool("raw", recOptions.PrintRaw, "receive option: print raw pulse data")
	var vClean = flag.Bool("clean", recOptions.PrintClean, "receive option: print cleaned up pulse data")

	var vMessage = flag.Bool("msg", false, "print message")
	var vBytes = flag.Bool("bytes", false, "print message as bytes")
	var vDiff = flag.Bool("diff", false, "print difference from previous")
	var vConfig = flag.Bool("config", false, "print decoded configuration")

	flag.Parse()

	if *vHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *vIrInput == "" {
		fmt.Printf("please set the device or file to read from")
		os.Exit(1)
	}

	slog.New(slog.NewTextHandler(os.Stdout, utils.SetLoggerOpts(*vLogLevel)))

	recOptions.Device = *vDevice
	recOptions.PrintRaw = *vRaw
	recOptions.PrintClean = *vClean

	options := Options{
		PrintBytes:   *vBytes,
		PrintDiff:    *vDiff,
		PrintConfig:  *vConfig,
		PrintMessage: *vMessage,
	}

	err := codec.StartReceiver(*vIrInput, messageHandler(&options), recOptions)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
