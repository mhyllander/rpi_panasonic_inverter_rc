package codec

import (
	"bufio"
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/ioctl"
	"strings"
	"time"
)

type ReceiverOptions struct {
	Device     bool
	PrintRaw   bool
	PrintClean bool
}

var receiveCommands chan string

func SuspendReceiver() {
	slog.Debug("suspending IR receiver")
	if receiveCommands != nil {
		receiveCommands <- "suspend"
	}
}

func ResumeReceiver() {
	slog.Debug("resuming IR receiver")
	if receiveCommands != nil {
		receiveCommands <- "resume"
	}
}

// ensure there are reasonable defaults
func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{Device: true}
}

func processMessages(messageStream chan *Message, processor func(*Message), options *ReceiverOptions) {
	slog.Debug("starting Message processor")
	for {
		msg := <-messageStream
		processor(msg)
	}
}

func processLircRawData(lircStream chan uint32, messageStream chan *Message, options *ReceiverOptions) {
	slog.Debug("starting LIRC processor")
	lircData := make([]uint32, 0, 10240)
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

func openFile(file string, options *ReceiverOptions) (*os.File, error) {
	slog.Debug("opening IR input", "file", file)
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	if options.Device {
		ioctl.SetLircReceiveMode(f)
	}

	return f, nil
}

func startReader(f *os.File, readStream chan []byte) {
	slog.Debug("starting IR reader")
	reader := bufio.NewReader(f)
	for {
		// create a new buffer for each read so it can be sent on the channel
		// and not be overwritten by subsequent reads
		readBuffer := make([]byte, 1024)
		n, err := reader.Read(readBuffer)
		if err != nil {
			if strings.Contains(err.Error(), "file already closed") {
				slog.Debug("IR input is closed, reader stopped")
			} else {
				slog.Error("failed to read from IR input", "error", err)
			}
			return
		}
		readStream <- readBuffer[:n]
	}
}

func StartIrReceiver(file string, messageHandler func(*Message), options *ReceiverOptions) error {
	slog.Debug("starting IR receiver")
	receiveCommands = make(chan string)

	f, err := openFile(file, options)
	if err != nil {
		return err
	}
	defer f.Close()

	messageStream := make(chan *Message)
	lircStream := make(chan uint32)
	readStream := make(chan []byte)
	go processMessages(messageStream, messageHandler, options)
	go processLircRawData(lircStream, messageStream, options)
	go startReader(f, readStream)

	for {
		select {
		case bytes := <-readStream:
			if len(bytes)%4 != 0 {
				slog.Debug("didn't get even 4 bytes matching uint32")
			}
			lircData := convertRawToLirc(bytes)
			for _, d := range lircData {
				lircStream <- d
			}
		case cmd := <-receiveCommands:
			switch cmd {
			case "suspend":
				// Suspend is sent before sending an IR message. We close the input
				// so that we don't receive our own message. This should cause the
				// startReader func to return.
				if f != nil {
					f.Close()
				}
			case "resume":
				// Resume is sent after sending an IR message. We a second and then open
				// the input and start a new reader.
				time.Sleep(time.Second)
				f, err = openFile(file, options)
				if err != nil {
					slog.Error("failed to open IR input", "file", file, "err", err)
					break
				}
				go startReader(f, readStream)
			}
		}

		if !options.Device {
			time.Sleep(100 * time.Millisecond)
			break
		}
	}
	return nil
}
