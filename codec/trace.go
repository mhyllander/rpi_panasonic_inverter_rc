package codec

import "fmt"

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

func PrintConfig(c *IrConfig) {
	fmt.Printf("power=%d mode=%d powerful=%d quiet=%d temp=%d fan=%d vpos=%d hpos=%d\n",
		c.Power, c.Mode, c.Powerful, c.Quiet, c.Temperature, c.FanSpeed, c.VentVertical, c.VentHorizontal)

	fmt.Printf(
		"timer_on: enabled=%d time=%s,  timer_off: enabled=%d time=%s,  clock: time=%s\n",
		c.TimerOnEnabled, c.TimerOn, c.TimerOffEnabled, c.TimerOff, c.Clock)
}

func PrintParams(msg *Message) {
	c := NewIrConfig(msg)
	PrintConfig(c)
}

func PrintMessage(msg *Message) {
	t1, p1 := msg.Frame1.ToTraceString()
	t2, p2 := msg.Frame2.ToTraceString()

	fmt.Printf("Message as bit stream (first and least significant bit to the right)\n")
	fmt.Printf("   %s\n%d: %s\n", p1, 1, t1)
	fmt.Printf("   %s\n%d: %s\n", p2, 2, t2)
}

func PrintByteRepresentation(msg *Message) {
	fmt.Println("Byte representation:")
	fmt.Printf("  %d: %s\n", 1, msg.Frame1.ToByteString())
	fmt.Printf("  %d: %s\n", 2, msg.Frame2.ToByteString())
}
