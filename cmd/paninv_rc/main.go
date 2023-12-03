package main

import (
	"flag"
	"fmt"
	"os"
	"rpi_panasonic_inverter_rc/codec"
)

// type Options struct {
// 	Byte        bool
// 	Diff        bool
// 	Param       bool
// 	sendOptions *codec.SenderOptions
// }

func main() {
	var vFile = flag.String("file", "/dev/lirc-tx", "LIRC transmit socket")
	var vSock = flag.Bool("sock", false, "writing to a socket")
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

	ic := codec.NewIrConfig(nil)
	m := ic.ToMessage()
	m.Frame2.SetChecksum()
	lircData := m.ToLirc()
	// codec.PrintLircBuffer(lircData)

	flags := os.O_WRONLY
	if !*vSock {
		flags = flags | os.O_CREATE
	}

	f, err := os.OpenFile(*vFile, flags, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	b := lircData.ToBytes()
	n, err := f.Write(b)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("wrote %d of %d bytes\n", n, len(b))

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
