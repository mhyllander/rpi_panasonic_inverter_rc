package codec

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"rpi_panasonic_inverter_rc/codecbase"
	"rpi_panasonic_inverter_rc/ioctl"
)

type SenderOptions struct {
	Mode2         bool
	Device        bool
	Transmissions int
	Interval_ms   int
}

// ensure there are reasonable defaults
func NewSenderOptions() *SenderOptions {
	return &SenderOptions{Device: true, Transmissions: 1, Interval_ms: 20}
}

var (
	ErrSenderStopped = errors.New("IR sender is stopped")
	ErrSenderBusy    = errors.New("IR sender queue is full")
)

const sendQueueSize = 10

type IrSender struct {
	irOutputFile  string
	senderOptions SenderOptions
	receiver      *IrReceiver // suspended while sending, can be nil
	sendChannel   chan *RcConfig
	stopWait      sync.WaitGroup

	// protects stopped and closing of sendChannel
	stopMu  sync.RWMutex
	stopped bool
}

// StartIrSender starts a sender that sends the queued configs one at a time. If a receiver is given,
// it is suspended while sending.
func StartIrSender(irOutputFile string, senderOptions *SenderOptions, receiver *IrReceiver) *IrSender {
	sender := &IrSender{
		irOutputFile:  irOutputFile,
		senderOptions: *senderOptions,
		receiver:      receiver,
		sendChannel:   make(chan *RcConfig, sendQueueSize),
	}
	sender.stopWait.Add(1)
	go sender.processConfigs()
	return sender
}

// SendConfig queues a config for sending. It never blocks, but fails if the queue is full or the sender is stopped.
func (sender *IrSender) SendConfig(sendRc *RcConfig) error {
	sender.stopMu.RLock()
	defer sender.stopMu.RUnlock()

	if sender.stopped {
		return ErrSenderStopped
	}

	select {
	case sender.sendChannel <- sendRc:
		return nil
	default:
		return ErrSenderBusy
	}
}

// Stop sends the already queued configs and then stops the sender. Further calls to SendConfig fail.
func (sender *IrSender) Stop() {
	sender.stopMu.Lock()
	if !sender.stopped {
		sender.stopped = true
		close(sender.sendChannel)
	}
	sender.stopMu.Unlock()

	sender.stopWait.Wait()
}

// Send the queued configs.
func (sender *IrSender) processConfigs() {
	defer sender.stopWait.Done()
	for sendRc := range sender.sendChannel {
		sender.send(sendRc)
	}
}

// Actually send a config. This is a separate function so we can make use of defer to close resources after sending.
func (sender *IrSender) send(sendRc *RcConfig) {
	// suspend the receiver while sending
	if sender.receiver != nil {
		sender.receiver.Suspend()
		defer sender.receiver.Resume()
	}

	f, err := sender.openIrOutputFile()
	if err != nil {
		slog.Error("failed to open IR output file", "err", err)
		return
	}
	defer f.Close()

	sendRc.LogConfigAndChecksum("sending config", "")
	err = SendIrConfig(sendRc, f, &sender.senderOptions)
	if err != nil {
		slog.Error("failed to send current config", "err", err)
	}
}

// Open file or device for sending IR
func (sender *IrSender) openIrOutputFile() (*os.File, error) {
	flags := os.O_RDWR
	if !sender.senderOptions.Device {
		flags = flags | os.O_CREATE
	}
	return os.OpenFile(sender.irOutputFile, flags, 0644)
}

// Prepare for sending by zeroing out all Mode2 type bits.
func stripMode2Types(lircData *LircBuffer) {
	ln := len(lircData.buf)
	for i := 0; i < ln; i++ {
		lircData.buf[i] = lircData.buf[i] & codecbase.L_LIRC_VALUE_MASK
	}
}

// When transmitting data over IR, the LIRC transmit socket expects a series of uint32 consisting of pulses and spaces.
// The data must start and end with a pulse, so there must be an odd number of uint32. In addition, no Mode2 bits
// should be set in the pulses and spaces (i.e. the send format is different from the receive format).
//
// The function can also write Mode2 to a file, and in that case it will keep the Mode2 types. This can be used to
// test that the output can be read as input, parsed correctly, and yield the original results.
func SendIrConfig(rc *RcConfig, f *os.File, options *SenderOptions) error {
	if options.Mode2 {
		s := rc.ConvertToLircData().ToMode2Lirc()
		s2 := strings.Join(s, " ")
		_, err := f.WriteString(s2)
		if err != nil {
			return err
		}
		slog.Debug("wrote mode2", "ints", len(s))
	} else {
		lircData := rc.ConvertToLircData()
		stripMode2Types(lircData)
		b := lircData.ToBytes()
		if options.Device {
			ioctl.SetLircSendMode(f)
		}
		for i := 0; i < options.Transmissions; i++ {
			n, err := f.Write(b)
			if err != nil {
				return err
			}
			slog.Debug("wrote raw LIRC", "bytes", len(b), "written", n)
			if i < options.Transmissions-1 {
				time.Sleep(time.Duration(options.Interval_ms) * time.Millisecond)
			}
		}
	}
	return nil
}
