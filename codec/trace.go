package codec

import "fmt"

func PrintConfig(c IrConfig) {
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
