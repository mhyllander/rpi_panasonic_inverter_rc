package utils

import (
	"rpi_panasonic_inverter_rc/codec"
	"rpi_panasonic_inverter_rc/rcconst"
	"testing"
)

func TestAdjustingPowerSetting1(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(6, 0)
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(18, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(4, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not adjusted to Off before TimerOn")
	}
}

func TestAdjustingPowerSetting2(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(6, 0)
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(18, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(12, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not adjusted to On between TimerOn and TimerOff")
	}
}

func TestAdjustingPowerSetting3(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(6, 0)
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(18, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(20, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not adjusted to Off after TimerOff")
	}
}

func TestAdjustingPowerSetting4(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(6, 0)
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(18, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(4, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not adjusted to On before TimerOff")
	}
}

func TestAdjustingPowerSetting5(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(6, 0)
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(18, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(12, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not adjusted to Off between TimerOff and TimerOn")
	}
}

func TestAdjustingPowerSetting6(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(6, 0)
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(18, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(20, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not adjusted to Off after TimerOn")
	}
}

func TestAdjustingPowerSetting7(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(6, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not adjusted to Off before TimerOn")
	}
}

func TestAdjustingPowerSetting8(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(18, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not adjusted to On after TimerOn")
	}
}

func TestAdjustingPowerSetting9(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(6, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not adjusted to On before TimerOff")
	}
}

func TestAdjustingPowerSetting10(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("", rc, dbRc, codec.NewTime(18, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not adjusted to Off after TimerOff")
	}
}

func TestAdjustingPowerSetting11(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("on", rc, dbRc, codec.NewTime(6, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not overridden to On before TimerOn")
	}
}

func TestAdjustingPowerSetting12(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOn = rcconst.C_Timer_Enabled
	dbRc.TimerOnTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("off", rc, dbRc, codec.NewTime(18, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not overridden to Off after TimerOn")
	}
}

func TestAdjustingPowerSetting13(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_On
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("off", rc, dbRc, codec.NewTime(6, 0))

	if rc.Power != rcconst.C_Power_Off {
		t.Fatal("Power was not overridden to Off before TimerOff")
	}
}

func TestAdjustingPowerSetting14(t *testing.T) {
	dbRc := codec.NewRcConfig()
	dbRc.Power = rcconst.C_Power_Off
	dbRc.TimerOff = rcconst.C_Timer_Enabled
	dbRc.TimerOffTime = codec.NewTime(12, 0)

	rc := dbRc.CopyForSending()

	setPower("on", rc, dbRc, codec.NewTime(18, 0))

	if rc.Power != rcconst.C_Power_On {
		t.Fatal("Power was not overridden to On after TimerOff")
	}
}
