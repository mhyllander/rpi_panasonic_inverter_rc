package codec

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

type SenderOptions struct {
	Mode2  bool
	Device bool
	Raw    bool
	Clean  bool
	Trace  bool
}

func setLircSendMode(f *os.File) {
	features, err := unix.IoctlGetUint32(int(f.Fd()), l_LIRC_GET_FEATURES)
	if err != nil {
		fmt.Println("ioctl error", err)
	}
	if features&l_LIRC_CAN_SEND_PULSE != 0 {
		mode := l_LIRC_MODE_PULSE
		err = unix.IoctlSetPointerInt(int(f.Fd()), l_LIRC_SET_SEND_MODE, mode)
		if err != nil {
			fmt.Println("ioctl error", err)
		}
	}
	if features&l_LIRC_CAN_SET_SEND_CARRIER != 0 {
		carrier := 38
		err = unix.IoctlSetPointerInt(int(f.Fd()), l_LIRC_SET_SEND_CARRIER, carrier)
		if err != nil {
			fmt.Println("ioctl error", err)
		}
	}
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
func SendIr(ic *IrConfig, f *os.File, options *SenderOptions) error {
	if options.Mode2 {
		s := ic.ConvertToMode2LircData()
		s2 := strings.Join(s, " ")
		_, err := f.WriteString(s2)
		if err != nil {
			return err
		}
		if options.Trace {
			fmt.Printf("wrote %d ints\n", len(s))
		}
	} else {
		licrData := ic.ConvertToLircData()
		stripMode2Types(licrData)
		b := licrData.ToBytes()
		if options.Device {
			setLircSendMode(f)
		}
		n, err := f.Write(b)
		if err != nil {
			return err
		}
		if options.Trace {
			fmt.Printf("wrote %d of %d bytes\n", n, len(b))
		}
	}
	return nil
}
