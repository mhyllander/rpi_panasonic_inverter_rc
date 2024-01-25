package codec

import (
	"testing"
)

func TestConversions(t *testing.T) {
	c1 := NewIrConfig(nil)

	m1 := c1.ToMessage()
	// the checksum is never set automatically,
	// since the clock time may be set right before sending the message
	m1.Frame2.SetChecksum()

	if !m1.Frame1.VerifyChecksum() {
		t.Fatalf("m1.Frame1 checksum wrong")
	}
	if !m1.Frame2.VerifyChecksum() {
		t.Fatalf("m1.Frame2 checksum wrong")
	}

	c2 := NewIrConfig(m1)

	if *c2 != *c1 {
		t.Fatalf("c2 config not equal to original c1")
	}

	m2 := c2.ToMessage()
	m2.Frame2.SetChecksum()

	if !m2.Frame1.VerifyChecksum() {
		t.Fatalf("m2.Frame1 checksum wrong")
	}
	if !m2.Frame2.VerifyChecksum() {
		t.Fatalf("m2.Frame2 checksum wrong")
	}

	if !m1.Frame1.Equal(m2.Frame1) {
		t.Fatalf("m2.Frame1 config not equal to m1.Frame1")
	}
	if !m1.Frame2.Equal(m2.Frame2) {
		t.Fatalf("m2.Frame2 config not equal to m1.Frame2")
	}
}
