package ioctl

import (
	"log/slog"
	"os"

	"golang.org/x/sys/unix"
)

const (
	// ioctl constant from /usr/include/linux/lirc.h
	ioctl_LIRC_GET_FEATURES         = uint(0x80046900)
	ioctl_LIRC_CAN_REC_MODE2        = uint32(0x00040000)
	ioctl_LIRC_CAN_SEND_PULSE       = uint32(0x00000002)
	ioctl_LIRC_CAN_SET_SEND_CARRIER = uint32(0x00000100)

	ioctl_LIRC_SET_REC_TIMEOUT_REPORTS = uint(0x40046919)

	ioctl_LIRC_SET_SEND_CARRIER = uint(0x40046913)

	ioctl_LIRC_GET_SEND_MODE = uint(0x80046901)
	ioctl_LIRC_SET_SEND_MODE = uint(0x40046911)
	ioctl_LIRC_MODE_PULSE    = 0x00000002
)

func SetLircReceiveMode(f *os.File) {
	features, err := unix.IoctlGetUint32(int(f.Fd()), ioctl_LIRC_GET_FEATURES)
	if err != nil {
		slog.Error("ioctl error", "error", err)
	}
	if features&ioctl_LIRC_CAN_REC_MODE2 == 0 {
		slog.Error("device can't receive mode2")
	}
	enabled := 1
	err = unix.IoctlSetPointerInt(int(f.Fd()), ioctl_LIRC_SET_REC_TIMEOUT_REPORTS, enabled)
	if err != nil {
		slog.Error("ioctl error enabling timeout reports", "error", err)
	}
}

func SetLircSendMode(f *os.File) {
	features, err := unix.IoctlGetUint32(int(f.Fd()), ioctl_LIRC_GET_FEATURES)
	if err != nil {
		slog.Error("ioctl error getting lirc features", "error", err)
	}
	if features&ioctl_LIRC_CAN_SEND_PULSE != 0 {
		mode := ioctl_LIRC_MODE_PULSE
		err = unix.IoctlSetPointerInt(int(f.Fd()), ioctl_LIRC_SET_SEND_MODE, mode)
		if err != nil {
			slog.Error("ioctl error setting send mode pulse", "error", err)
		}
	} else {
		slog.Debug("ioctl doesn't support setting mode pulse")
	}
	if features&ioctl_LIRC_CAN_SET_SEND_CARRIER != 0 {
		carrier := 38000
		err = unix.IoctlSetPointerInt(int(f.Fd()), ioctl_LIRC_SET_SEND_CARRIER, carrier)
		if err != nil {
			slog.Error("ioctl error setting send carrier", "error", err)
		}
	} else {
		slog.Debug("ioctl doesn't support setting send carrier")
	}
}
