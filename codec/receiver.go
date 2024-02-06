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

type command struct {
	cmd     string
	confirm chan<- struct{}
}

var receiveCommands chan command

func SuspendReceiver(confirmCommand chan<- struct{}) {
	slog.Debug("suspending IR receiver")
	if receiveCommands != nil {
		receiveCommands <- command{"suspend", confirmCommand}
	}
}

func ResumeReceiver(confirmCommand chan<- struct{}) {
	slog.Debug("resuming IR receiver")
	if receiveCommands != nil {
		receiveCommands <- command{"resume", confirmCommand}
	}
}

func QuitReceiver(confirmCommand chan<- struct{}) {
	slog.Debug("quiting IR receiver")
	if receiveCommands != nil {
		receiveCommands <- command{"quit", confirmCommand}
	}
}

// ensure there are reasonable defaults
func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{Device: true}
}

func processMessages(messageStream <-chan *Message, processor func(*Message), options *ReceiverOptions) {
	slog.Debug("starting Message processor")
	for {
		msg, ok := <-messageStream
		if !ok {
			slog.Debug("messageStream was closed")
			return
		}
		processor(msg)
	}
}

func processLircRawData(lircStream <-chan uint32, messageStream chan<- *Message, options *ReceiverOptions) {
	slog.Debug("starting LIRC processor")
	defer close(messageStream)
	lircData := make([]uint32, 0, 10240)
	for {
		d, ok := <-lircStream
		if !ok {
			slog.Debug("lircStream was closed")
			return
		}
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
		case PARSE_OK:
			// send message
			messageStream <- msg
		case PARSE_NOT_ENOUGH_DATA:
		case PARSE_END_OF_DATA:
		default:
			slog.Debug("problem during parsing", "state", state)
		}
		// copy remaining data to start of lircData
		lircData = lircData[:len(remainingData)]
		copy(lircData, remainingData)
	}
}

func startReader(f *os.File, lircStream chan<- uint32) {
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
		bytes := readBuffer[:n]
		if len(bytes)%4 != 0 {
			slog.Debug("didn't get even 4 bytes matching uint32")
		}
		lircData := convertRawToLirc(bytes)
		for _, d := range lircData {
			lircStream <- d
		}
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

func RunIrReceiver(file string, messageHandler func(*Message), options *ReceiverOptions) error {
	slog.Debug("starting IR receiver")

	f, err := openFile(file, options)
	if err != nil {
		return err
	}
	defer f.Close()

	receiveCommands = make(chan command)
	defer func() {
		close(receiveCommands)
		receiveCommands = nil
	}()

	messageStream := make(chan *Message)
	lircStream := make(chan uint32)

	// start the processing pipeline
	go processMessages(messageStream, messageHandler, options)
	go processLircRawData(lircStream, messageStream, options)
	// closing lircStream will close the processing pipeline
	defer close(lircStream)

	// the reader can be started and stopped independently (by closing f)
	go startReader(f, lircStream)

	for {
		cmd := <-receiveCommands
		switch cmd.cmd {
		case "suspend":
			// Suspend is sent before sending an IR message. We close the input
			// so that we don't receive our own message. This should cause the
			// startReader func to return.
			if f != nil {
				f.Close()
			}
		case "resume":
			// Resume is sent after sending an IR message. Wait a moment and then open
			// the input and start a new reader.
			time.Sleep(2 * time.Second)
			f, err = openFile(file, options)
			if err != nil {
				slog.Error("failed to open IR input", "file", file, "err", err)
				break
			}
			go startReader(f, lircStream)
		case "quit":
			// Quit is sent to stop the receiver completely. All channels and files will be closed,
			// and goroutines will exit.
			return nil
		}
		// send confirmation if requested
		if cmd.confirm != nil {
			cmd.confirm <- struct{}{}
		}
	}
}
