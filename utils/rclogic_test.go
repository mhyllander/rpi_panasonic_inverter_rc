package utils

import (
	"rpi_panasonic_inverter_rc/codec"
	"testing"
)

func TestAdjustingPowerSetting1(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_On
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(6, 0)
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(18, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(4, 0))

	if ic.Power != codec.C_Power_Off {
		t.Fatal("Power was not adjusted to Off before TimerOn")
	}
}

func TestAdjustingPowerSetting2(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_Off
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(6, 0)
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(18, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(12, 0))

	if ic.Power != codec.C_Power_On {
		t.Fatal("Power was not adjusted to On between TimerOn and TimerOff")
	}
}

func TestAdjustingPowerSetting3(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_On
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(6, 0)
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(18, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(20, 0))

	if ic.Power != codec.C_Power_Off {
		t.Fatal("Power was not adjusted to Off after TimerOff")
	}
}

func TestAdjustingPowerSetting4(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_Off
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(6, 0)
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(18, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(4, 0))

	if ic.Power != codec.C_Power_On {
		t.Fatal("Power was not adjusted to On before TimerOff")
	}
}

func TestAdjustingPowerSetting5(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_On
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(6, 0)
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(18, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(12, 0))

	if ic.Power != codec.C_Power_Off {
		t.Fatal("Power was not adjusted to Off between TimerOff and TimerOn")
	}
}

func TestAdjustingPowerSetting6(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_Off
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(6, 0)
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(18, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(20, 0))

	if ic.Power != codec.C_Power_On {
		t.Fatal("Power was not adjusted to Off after TimerOn")
	}
}

func TestAdjustingPowerSetting7(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_On
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(12, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(6, 0))

	if ic.Power != codec.C_Power_Off {
		t.Fatal("Power was not adjusted to Off before TimerOn")
	}
}

func TestAdjustingPowerSetting8(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_Off
	dbIc.TimerOn = codec.C_Timer_Enabled
	dbIc.TimerOnTime = codec.NewTime(12, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(18, 0))

	if ic.Power != codec.C_Power_On {
		t.Fatal("Power was not adjusted to On after TimerOn")
	}
}

func TestAdjustingPowerSetting9(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_Off
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(12, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(6, 0))

	if ic.Power != codec.C_Power_On {
		t.Fatal("Power was not adjusted to On before TimerOff")
	}
}

func TestAdjustingPowerSetting10(t *testing.T) {
	dbIc := codec.NewIrConfig(nil)
	dbIc.Power = codec.C_Power_On
	dbIc.TimerOff = codec.C_Timer_Enabled
	dbIc.TimerOffTime = codec.NewTime(12, 0)

	ic := dbIc.CopyForSending()

	setPower("", ic, dbIc, codec.NewTime(18, 0))

	if ic.Power != codec.C_Power_Off {
		t.Fatal("Power was not adjusted to Off after TimerOff")
	}
}
