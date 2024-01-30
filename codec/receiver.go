package codec

import (
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/ioctl"
	"time"
)

type receiverOptions struct {
	Device     bool
	PrintRaw   bool
	PrintClean bool
}

// ensure there are reasonable defaults
func NewReceiverOptions() *receiverOptions {
	return &receiverOptions{Device: true}
}

func processMessages(messageStream chan *Message, processor func(*Message), options *receiverOptions) {
	for {
		msg := <-messageStream
		processor(msg)
	}
}

func newBuffer() []uint32 {
	return make([]uint32, 0, 10240)
}

func processLircRawData(lircStream chan uint32, messageStream chan *Message, options *receiverOptions) {
	lircData := newBuffer()
	for {
		d := <-lircStream
		if options.PrintRaw {
			printLircData("raw", d)
		}
		keep, d := filterLircAsPanasonic(d)
		if !keep {
			continue
		}
		if options.PrintClean {
			printLircData("clean", d)
		}
		lircData = append(lircData, d)
		msg, remainingData, state := readPanasonicMessage(lircData, options)
		switch state.status {
		case PARSE_NOT_ENOUGH_DATA:
		case PARSE_END_OF_DATA:
		case PARSE_OK:
			// send message
			messageStream <- msg
		default:
			slog.Debug("problem during parsing", "state", state)
		}
		// copy remaining data to start of lircData
		lircData = lircData[:len(remainingData)]
		copy(lircData, remainingData)
	}
}

func StartIrReceiver(file string, messageHandler func(*Message), options *receiverOptions) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if options.Device {
		ioctl.SetLircReceiveMode(f)
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
		if n%4 != 0 {
			slog.Debug("didn't get even 4 bytes matching uint32")
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
