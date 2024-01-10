package codec

import (
	"fmt"
	"log/slog"
)

func printLircData(label string, d uint32) {
	v := d & l_LIRC_VALUE_MASK
	fmt.Printf("%s\t", label)
	switch d & l_LIRC_MODE2_MASK {
	case l_LIRC_MODE2_SPACE:
		fmt.Printf("space\t%d\n", v)
	case l_LIRC_MODE2_PULSE:
		fmt.Printf("pulse\t%d\n", v)
	case l_LIRC_MODE2_FREQUENCY:
		fmt.Printf("frequencyt%d\n", v)
	case l_LIRC_MODE2_TIMEOUT:
		fmt.Printf("timeout\t%d\n", v)
	case l_LIRC_MODE2_OVERFLOW:
		fmt.Printf("overflow\t%d\n", v)
	}
}

func PrintLircBuffer(b *LircBuffer) {
	for _, code := range b.buf {
		printLircData("LircBuffer", code)
	}
}

func PrintMessage(msg *Message) {
	t1, p1 := msg.Frame1.ToVerboseString()
	t2, p2 := msg.Frame2.ToVerboseString()

	fmt.Printf("Message as bit stream (first and least significant bit to the right)\n")
	fmt.Printf("   %s\n%d: %s\n", p1, 1, t1)
	fmt.Printf("   %s\n%d: %s\n", p2, 2, t2)
}

func PrintByteRepresentation(msg *Message) {
	fmt.Println("Byte representation:")
	fmt.Printf("  %d: %s\n", 1, msg.Frame1.ToByteString())
	fmt.Printf("  %d: %s\n", 2, msg.Frame2.ToByteString())
}

func toOnOffString(v uint) string {
	switch v {
	case p_PANASONIC_ENABLED:
		return "on"
	case p_PANASONIC_DISABLED:
		return "off"
	}
	return "<bad value>"
}

func toModeString(mode uint) string {
	switch mode {
	case C_Mode_Auto:
		return "auto"
	case C_Mode_Cool:
		return "cool"
	case C_Mode_Heat:
		return "heat"
	case C_Mode_Dry:
		return "dry"
	}
	return "<bad mode>"
}

func toFanSpeedString(fan uint) string {
	switch fan {
	case C_FanSpeed_Auto:
		return "auto"
	case C_FanSpeed_Lowest:
		return "lowest"
	case C_FanSpeed_Low:
		return "low"
	case C_FanSpeed_Middle:
		return "middle"
	case C_FanSpeed_High:
		return "high"
	case C_FanSpeed_Highest:
		return "highest"
	}
	return "<bad fan speed>"
}

func toVentVerticalString(vert uint) string {
	switch vert {
	case C_VentVertical_Auto:
		return "auto"
	case C_VentVertical_Lowest:
		return "lowest"
	case C_VentVertical_Low:
		return "low"
	case C_VentVertical_Middle:
		return "middle"
	case C_VentVertical_High:
		return "high"
	case C_VentVertical_Highest:
		return "highest"
	}
	return "<bad vent vertical>"
}

func toVentHorizontalString(horiz uint) string {
	switch horiz {
	case C_VentHorizontal_Auto:
		return "auto"
	case C_VentHorizontal_FarLeft:
		return "farleft"
	case C_VentHorizontal_Left:
		return "left"
	case C_VentHorizontal_Middle:
		return "middle"
	case C_VentHorizontal_Right:
		return "right"
	case C_VentHorizontal_FarRight:
		return "farright"
	}
	return "<bad vent horizontal>"
}

func PrintConfigAndChecksum(c *IrConfig, checksumStatus string) {
	fmt.Printf("power=%s(%d) mode=%s(%d) powerful=%s(%d) quiet=%s(%d)\n",
		toOnOffString(c.Power), c.Power,
		toModeString(c.Mode), c.Mode,
		toOnOffString(c.Powerful), c.Powerful,
		toOnOffString(c.Quiet), c.Quiet)

	fmt.Printf("temp=%d fan=%s(%d) vent.vert=%s(%d) vent.horiz=%s(%d)\n",
		c.Temperature,
		toFanSpeedString(c.FanSpeed), c.FanSpeed,
		toVentVerticalString(c.VentVertical), c.VentVertical,
		toVentHorizontalString(c.VentHorizontal), c.VentHorizontal)

	fmt.Printf(
		"timer_on=%s(%d) timer_on.time=%s,  timer_off=%s(%d) timer_off.time=%s,  clock: time=%s\n",
		toOnOffString(c.TimerOnEnabled), c.TimerOnEnabled, c.TimerOn, toOnOffString(c.TimerOffEnabled), c.TimerOffEnabled, c.TimerOff, c.Clock)

	if checksumStatus != "" {
		fmt.Printf("checksum: %s\n", checksumStatus)
	}
}

func LogConfigAndChecksum(c *IrConfig, checksumStatus string) {
	slog.Info("config",
		"power", toOnOffString(c.Power),
		"mode", toModeString(c.Mode),
		"powerful", toOnOffString(c.Powerful),
		"quiet", toOnOffString(c.Quiet),
		"temp", c.Temperature,
		"fan", toFanSpeedString(c.FanSpeed),
		"vent.vert", toVentVerticalString(c.VentVertical),
		"vent.horiz", toVentHorizontalString(c.VentHorizontal),
		"timer_on", toOnOffString(c.TimerOnEnabled),
		"timer_on.time", c.TimerOn,
		"timer_off", toOnOffString(c.TimerOffEnabled),
		"timer_off.time", c.TimerOff,
		"clock", c.Clock,
		"checksum", checksumStatus,
	)
}
