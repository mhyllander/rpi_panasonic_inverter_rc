# Hardware and Circuits

## Various tutorials

1. https://www.digikey.se/en/maker/tutorials/2021/how-to-send-and-receive-ir-signals-with-a-raspberry-pi
2. https://devkimchi.com/2020/08/12/turning-raspberry-pi-into-remote-controller/
3. https://blog.gordonturner.com/2020/05/31/raspberry-pi-ir-receiver/
4. https://blog.gordonturner.com/2020/06/10/raspberry-pi-ir-transmitter/
5. https://github.com/gordonturner/ControlKit/blob/master/Raspbian%20Setup%20and%20Configure%20IR.md

## List of components

TSOP38238 IR receiver, 38 kHz, 940 nm<br>
(Supply voltage: 2.5 V to 5.5 V)<br>
Datasheet https://www.vishay.com/docs/82491/tsop382.pdf

TSAL6200 IR LED, 940 nm<br>
Datasheet https://www.vishay.com/docs/81010/tsal6200.pdf<br>
(I used two of these, which is probably not necessary.)

NPN transistor 2N2222A<br>
Datasheet  https://components101.com/transistors/2n2222a-pinout-equivalent-datasheet

Resistors: 10 kΩ, 22 Ω

Resistors code calculator: https://resistorcolorcodecalc.com

## Connecting the Raspberry Pi Zero 2 W

https://www.raspberrypi.com/documentation/computers/raspberry-pi.html#gpio-and-the-40-pin-header

https://pinout.xyz

![PCB circuit](./pcb_circuit.png)

[pcb_circuit.drawio](./pcb_circuit.drawio)

The 22 Ω resistor was calculated as follows, with two LEDs in series:

* IR LED TSAL6200
  * V<sub>forward</sub> = 1,35 V
  * I<sub>max</sub> = 100 mA
* NPN 2N2222
  * V<sub>ce,sat</sub> = 0,3 V
  * I<sub>ce,max</sub> = 800 mA

I<sub>max</sub> for the LEDs is the limiting current. Voltage falls over each of the three components:

V<sub>R</sub> = 5,0 V - 2\*V<sub>forward</sub> - V<sub>ce,sat</sub> = 5,0 - 2\*1,35 - 0,3 V = 2,0 V

R = V<sub>R</sub> / I<sub>max</sub> = 2,0 V / 0,1 A = 20,0 Ω

The next biggest resistor was 22 Ω.

## Pictures

<img src="pizero_bottom.jpg" width="400">
<img src="pizero_top.jpg" width="400">
<img src="breadboard.jpg" width="400">
<img src="pizero_with_breadboard.jpg" width="400">
<img src="pcb1.jpg" width="400">
<img src="pcb2.jpg" width="400">
<img src="pizero_with_pcb.jpg" width="400">
