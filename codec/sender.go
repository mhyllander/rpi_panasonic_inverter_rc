package codec

import (
	"log/slog"
	"os"
	"rpi_panasonic_inverter_rc/ioctl"
	"strings"
	"time"
)

type senderOptions struct {
	Mode2         bool
	Device        bool
	Transmissions int
	Interval_ms   int
}

// ensure there are reasonable defaults
func NewSenderOptions() *senderOptions {
	return &senderOptions{Device: true, Transmissions: 4, Interval_ms: 20}
}

func stripMode2Types(licrData *LircBuffer) {
	ln := len(licrData.buf)
	for i := 0; i < ln; i++ {
		licrData.buf[i] = licrData.buf[i] & l_LIRC_VALUE_MASK
	}
}

// When transmitting data over IR, the LIRC transmit socket expects a series of uint32 consisting of pulses and spaces.
// The data must start and end with a pulse, so there must be an odd number of uint32. In addition, no mode2 bits
// should be set in the pulses and spaces (i.e. the send format is different from the receive format).
func SendIr(ic *IrConfig, f *os.File, options *senderOptions) error {
	if options.Mode2 {
		s := ic.ConvertToMode2LircData()
		s2 := strings.Join(s, " ")
		_, err := f.WriteString(s2)
		if err != nil {
			return err
		}
		slog.Debug("wrote mode2", "ints", len(s))
	} else {
		licrData := ic.ConvertToLircData()
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
