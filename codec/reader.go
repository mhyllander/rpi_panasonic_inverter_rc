package codec

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

type ReaderOptions struct {
	Socket bool
	Raw    bool
	Clean  bool
	Trace  bool
	Byte   bool
	Diff   bool
	Param  bool
}

func printLircData(label string, d uint32) {
	v := d & LIRC_VALUE_MASK
	fmt.Printf("%s\t", label)
	switch d & LIRC_MODE2_MASK {
	case LIRC_MODE2_SPACE:
		fmt.Printf("space\t%d\n", v)
	case LIRC_MODE2_PULSE:
		fmt.Printf("pulse\t%d\n", v)
	case LIRC_MODE2_FREQUENCY:
		fmt.Printf("frequencyt%d\n", v)
	case LIRC_MODE2_TIMEOUT:
		fmt.Printf("timeout\t%d\n", v)
	case LIRC_MODE2_OVERFLOW:
		fmt.Printf("overflow\t%d\n", v)
	}
}

func processMessages(messageStream chan *Message, processor func(*Message), options *ReaderOptions) {
	for {
		msg := <-messageStream
		processor(msg)
	}
}

func newBuffer() []uint32 {
	return make([]uint32, 0, 10240)
}

func processLircRawData(lircStream chan uint32, messageStream chan *Message, options *ReaderOptions) {
	lircData := newBuffer()
	for {
		d := <-lircStream
		if options.Raw {
			printLircData("raw", d)
		}
		keep, d := filterLircAsPanasonic(d)
		if !keep {
			continue
		}
		if options.Clean {
			printLircData("clean", d)
		}
		lircData = append(lircData, d)
		msg, remainingData, state := readPanasonicMessage(lircData, options)
		if state.status == PARSE_NOT_ENOUGH_DATA || state.status == PARSE_END_OF_DATA {
			continue
		}
		if state.status != PARSE_OK {
			fmt.Println(state)
			// failure recovery: create a new empty buffer if we can't parse the current one
			lircData = newBuffer()
			continue
		}
		// send message
		messageStream <- msg
		// copy remaining data to start of lircData
		lircData = lircData[:len(remainingData)]
		copy(lircData, remainingData)
	}
}

func disableTimeoutReports(f *os.File) {
	enabled := 0
	err := unix.IoctlSetPointerInt(int(f.Fd()), LIRC_SET_REC_TIMEOUT_REPORTS, enabled)
	if err != nil {
		fmt.Println("ioctl error", err)
	}
}

func StartReader(file string, processor func(*Message), options *ReaderOptions) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if options.Socket {
		disableTimeoutReports(f)
	}

	lircStream := make(chan uint32)
	messageStream := make(chan *Message)
	go processLircRawData(lircStream, messageStream, options)
	go processMessages(messageStream, processor, options)

	readBuffer := make([]byte, 10240)
	for {
		n, err := f.Read(readBuffer)
		if err != nil {
			return err
		}

		lircData := convertRawToLirc(readBuffer[:n])
		for _, d := range lircData {
			lircStream <- d
		}
		if !options.Socket {
			time.Sleep(100 * time.Millisecond)
			break
		}
	}
	return nil
}
