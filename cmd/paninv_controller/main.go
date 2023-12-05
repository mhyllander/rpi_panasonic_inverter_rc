package main

import "flag"

func main() {
	var vIrInput = flag.String("irin", "/dev/lirc-rx", "LIRC receive device ")
	var vIrOutput = flag.String("irout", "/dev/lirc-tx", "LIRC transmit device")
	var vIrDb = flag.String("db", "paninv.db", "SQLite database")
	var vHelp = flag.Bool("help", false, "print usage")
}
