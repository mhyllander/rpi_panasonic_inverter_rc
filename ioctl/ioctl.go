package ioctl

import (
	"fmt"
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
		fmt.Println("ioctl error", err)
	}
	if features&ioctl_LIRC_CAN_REC_MODE2 == 0 {
		fmt.Println("device can't receive mode2")
	}
	enabled := 1
	err = unix.IoctlSetPointerInt(int(f.Fd()), ioctl_LIRC_SET_REC_TIMEOUT_REPORTS, enabled)
	if err != nil {
		fmt.Println("ioctl error", err)
	}
}

func SetLircSendMode(f *os.File) {
	features, err := unix.IoctlGetUint32(int(f.Fd()), ioctl_LIRC_GET_FEATURES)
	if err != nil {
		fmt.Println("ioctl error", err)
	}
	if features&ioctl_LIRC_CAN_SEND_PULSE != 0 {
		mode := ioctl_LIRC_MODE_PULSE
		err = unix.IoctlSetPointerInt(int(f.Fd()), ioctl_LIRC_SET_SEND_MODE, mode)
		if err != nil {
			fmt.Println("ioctl error", err)
		}
	}
	if features&ioctl_LIRC_CAN_SET_SEND_CARRIER != 0 {
		carrier := 38000
		err = unix.IoctlSetPointerInt(int(f.Fd()), ioctl_LIRC_SET_SEND_CARRIER, carrier)
		if err != nil {
			fmt.Println("ioctl error", err)
		}
	}
}
