package codec

import (
	"bufio"
	"errors"
	"log/slog"
	"os"
	"time"

	"rpi_panasonic_inverter_rc/ioctl"
)

const (
	// Time to wait after a suspended receiver is resumed before the input is reopened.
	receiverResumeDelay = 2 * time.Second

	// Reopening the input is retried with an increasing delay between these limits.
	receiverMinRetryDelay = time.Second
	receiverMaxRetryDelay = 30 * time.Second
)

type ReceiverOptions struct {
	Device     bool
	PrintRaw   bool
	PrintClean bool
}

// ensure there are reasonable defaults
func NewReceiverOptions() *ReceiverOptions {
	return &ReceiverOptions{Device: true}
}

type receiverCommandKind int

const (
	commandSuspend receiverCommandKind = iota
	commandResume
	commandQuit
)

type receiverCommand struct {
	kind    receiverCommandKind
	confirm chan struct{}
}

// IrReceiver reads IR messages from a LIRC device or file. The receiver can be suspended while
// sending IR, so that it doesn't receive its own messages.
type IrReceiver struct {
	file           string
	messageHandler func(*Message)
	options        *ReceiverOptions

	commands chan receiverCommand
	reopen   chan struct{}

	// closed when Run returns, so that commands never block on a receiver that has stopped
	done chan struct{}
}

func NewIrReceiver(file string, messageHandler func(*Message), options *ReceiverOptions) *IrReceiver {
	return &IrReceiver{
		file:           file,
		messageHandler: messageHandler,
		options:        options,
		commands:       make(chan receiverCommand),
		reopen:         make(chan struct{}, 1),
		done:           make(chan struct{}),
	}
}

// Suspend closes the IR input and returns when it is closed. Does nothing if the receiver isn't running.
func (r *IrReceiver) Suspend() {
	slog.Debug("suspending IR receiver")
	r.sendCommand(commandSuspend)
}

// Resume reopens the IR input after a short delay, and keeps retrying until it succeeds. It returns
// immediately and does not wait for the input to be reopened.
func (r *IrReceiver) Resume() {
	slog.Debug("resuming IR receiver")
	r.sendCommand(commandResume)
}

// Quit stops the receiver, and makes Run return.
func (r *IrReceiver) Quit() {
	slog.Debug("quitting IR receiver")
	r.sendCommand(commandQuit)
}

// Send a command to the receiver and wait until it is handled. Returns early if the receiver has stopped
// (or has not been able to start).
func (r *IrReceiver) sendCommand(kind receiverCommandKind) {
	command := receiverCommand{kind: kind, confirm: make(chan struct{}, 1)}

	select {
	case r.commands <- command:
	case <-r.done:
		return
	}

	select {
	case <-command.confirm:
	case <-r.done:
	}
}

// Run starts the receiver and blocks until it quits or fails to open the input. It must only be called once.
func (r *IrReceiver) Run() error {
	defer close(r.done)
	slog.Debug("starting IR receiver")

	f, err := openFile(r.file, r.options)
	if err != nil {
		return err
	}

	// stop makes the processing pipeline and all readers exit
	stop := make(chan struct{})
	defer close(stop)

	messageStream := make(chan *Message)
	lircStream := make(chan uint32)
	go processMessages(messageStream, r.messageHandler)
	go processLircRawData(lircStream, messageStream, stop, r.options)

	// The reader can be started and stopped independently (by closing f).
	go startReader(f, lircStream, stop)

	var reopenTimer *time.Timer
	retryDelay := receiverMinRetryDelay
	paused := false

	closeInput := func() {
		if f != nil {
			f.Close()
			f = nil
		}
	}
	stopReopenTimer := func() {
		if reopenTimer != nil {
			reopenTimer.Stop()
		}
	}
	scheduleReopen := func(delay time.Duration) {
		stopReopenTimer()
		reopenTimer = time.AfterFunc(delay, func() {
			select {
			case r.reopen <- struct{}{}:
			default:
				// a reopen is already pending
			}
		})
	}
	defer func() {
		stopReopenTimer()
		closeInput()
	}()

	for {
		select {
		case command := <-r.commands:
			switch command.kind {
			case commandSuspend:
				// Suspend is sent before sending an IR message. We close the input
				// so that we don't receive our own message. This should cause the
				// startReader func to return.
				paused = true
				stopReopenTimer()
				closeInput()
			case commandResume:
				// Resume is sent after sending an IR message. Wait a moment and then
				// reopen the input and start a new reader.
				paused = false
				retryDelay = receiverMinRetryDelay
				scheduleReopen(receiverResumeDelay)
			case commandQuit:
				// Quit is sent to stop the receiver completely. All channels and files will be closed,
				// and goroutines will exit.
				command.confirm <- struct{}{}
				return nil
			}
			command.confirm <- struct{}{}

		case <-r.reopen:
			// Ignore stale timers: the receiver may have been suspended again or already reopened.
			if paused || f != nil {
				continue
			}

			f, err = openFile(r.file, r.options)
			if err != nil {
				slog.Error("failed to open IR input, will retry", "file", r.file, "retryIn", retryDelay, "err", err)
				scheduleReopen(retryDelay)
				retryDelay = min(retryDelay*2, receiverMaxRetryDelay)
				continue
			}
			go startReader(f, lircStream, stop)
		}
	}
}

func processMessages(messageStream <-chan *Message, processor func(*Message)) {
	slog.Debug("starting Message processor")
	for msg := range messageStream {
		processor(msg)
	}
	slog.Debug("messageStream was closed")
}

func processLircRawData(lircStream <-chan uint32, messageStream chan<- *Message, stop <-chan struct{}, options *ReceiverOptions) {
	slog.Debug("starting LIRC processor")
	defer close(messageStream)
	lircData := make([]uint32, 0, 10240)
	for {
		var d uint32
		select {
		case d = <-lircStream:
		case <-stop:
			slog.Debug("LIRC processor stopped")
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
			select {
			case messageStream <- msg:
			case <-stop:
				return
			}
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

func startReader(f *os.File, lircStream chan<- uint32, stop <-chan struct{}) {
	slog.Debug("starting IR reader")
	reader := bufio.NewReader(f)
	for {
		// create a new buffer for each read so it can be sent on the channel
		// and not be overwritten by subsequent reads
		readBuffer := make([]byte, 1024)
		n, err := reader.Read(readBuffer)
		if err != nil {
			if errors.Is(err, os.ErrClosed) {
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
			select {
			case lircStream <- d:
			case <-stop:
				return
			}
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
