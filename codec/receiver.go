package codec

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

type ReceiverOptions struct {
	Device bool
	Raw    bool
	Clean  bool
	Trace  bool
}

func processMessages(messageStream chan *Message, processor func(*Message), options *ReceiverOptions) {
	for {
		msg := <-messageStream
		processor(msg)
	}
}

func newBuffer() []uint32 {
	return make([]uint32, 0, 10240)
}

func processLircRawData(lircStream chan uint32, messageStream chan *Message, options *ReceiverOptions) {
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

func setLircReceiveMode(f *os.File) {
	features, err := unix.IoctlGetUint32(int(f.Fd()), l_LIRC_GET_FEATURES)
	if err != nil {
		fmt.Println("ioctl error", err)
	}
	if features&l_LIRC_CAN_REC_MODE2 == 0 {
		fmt.Println("device can't receive mode2")
	}
	enabled := 0
	err = unix.IoctlSetPointerInt(int(f.Fd()), l_LIRC_SET_REC_TIMEOUT_REPORTS, enabled)
	if err != nil {
		fmt.Println("ioctl error", err)
	}
}

func StartReceiver(file string, messageHandler func(*Message), options *ReceiverOptions) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if options.Device {
		setLircReceiveMode(f)
	}

	lircStream := make(chan uint32)
	messageStream := make(chan *Message)
	go processLircRawData(lircStream, messageStream, options)
	go processMessages(messageStream, messageHandler, options)

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
		if !options.Device {
			time.Sleep(100 * time.Millisecond)
			break
		}
	}
	return nil
}
