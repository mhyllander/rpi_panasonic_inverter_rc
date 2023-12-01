package main

import (
	"flag"
	"fmt"
	"os"
)

// type Options struct {
// 	Byte        bool
// 	Diff        bool
// 	Param       bool
// 	sendOptions *codec.SenderOptions
// }

func main() {
	var vFile = flag.String("sock", "/dev/lirc-tx", "LIRC transmit socket")
	// var vRaw = flag.Bool("raw", false, "print raw pulse data")
	// var vClean = flag.Bool("clean", false, "print cleaned up pulse data")
	// var vTrace = flag.Bool("trace", false, "print message trace")
	// var vByte = flag.Bool("byte", false, "print message as bytes")
	// var vDiff = flag.Bool("diff", false, "show difference from previous")
	// var vParam = flag.Bool("param", false, "show decoded params")
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

	// sendOptions := codec.SenderOptions{
	// 	Socket: *vSocket,
	// 	Raw:    *vRaw,
	// 	Clean:  *vClean,
	// 	Trace:  *vTrace,
	// }
	// options := Options{
	// 	Byte:        *vByte,
	// 	Diff:        *vDiff,
	// 	Param:       *vParam,
	// 	sendOptions: &sendOptions,
	// }

	// err := codec.Send(*vFile, messageHandler(&options), &sendOptions)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
}
