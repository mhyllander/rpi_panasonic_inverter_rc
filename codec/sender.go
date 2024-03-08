package codec

import (
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/common"
	"rpi_panasonic_inverter_rc/ioctl"
	"strings"
	"sync"
	"time"
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

type IrSender struct {
	irOutputFile  string
	senderOptions SenderOptions
	sendChannel   chan *RcConfig
	stopWait      sync.WaitGroup
}

func StartIrSender(irOutputFile string, senderOptions *SenderOptions) *IrSender {
	sender := &IrSender{irOutputFile, *senderOptions, make(chan *RcConfig, 10), sync.WaitGroup{}}
	sender.stopWait.Add(1)
	go sender.processConfigs()
	return sender
}

func (sender *IrSender) SendConfig(sendRc *RcConfig) {
	sender.sendChannel <- sendRc
}

func (sender *IrSender) Stop() {
	close(sender.sendChannel)
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
	var err error

	// suspend the receiver while sending
	confirmCommand := make(chan struct{})
	SuspendReceiver(confirmCommand)
	<-confirmCommand
	defer func() {
		ResumeReceiver(confirmCommand)
		<-confirmCommand
	}()

	f := sender.openIrOutputFile()
	if f == nil {
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
func (sender *IrSender) openIrOutputFile() *os.File {
	flags := os.O_RDWR
	if !sender.senderOptions.Device {
		flags = flags | os.O_CREATE
	}
	f, err := os.OpenFile(sender.irOutputFile, flags, 0644)
	if err != nil {
		slog.Error("failed to open IR output file", "err", err)
		return nil
	}
	return f
}

// Prepare for sending by zeroing out all Mode2 type bits.
func stripMode2Types(licrData *LircBuffer) {
	ln := len(licrData.buf)
	for i := 0; i < ln; i++ {
		licrData.buf[i] = licrData.buf[i] & common.L_LIRC_VALUE_MASK
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
		s := rc.ConvertToMode2LircData()
		s2 := strings.Join(s, " ")
		_, err := f.WriteString(s2)
		if err != nil {
			return err
		}
		slog.Debug("wrote mode2", "ints", len(s))
	} else {
		licrData := rc.ConvertToLircData()
		stripMode2Types(licrData)
		b := licrData.ToBytes()
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
