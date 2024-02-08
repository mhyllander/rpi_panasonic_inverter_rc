# Configuring the Raspberry Pi

## Some reference matterial

https://github.com/raspberrypi/linux/issues/2993#issuecomment-497420228

https://www.kernel.org/doc/html/v6.1/userspace-api/media/rc/lirc-dev.html

https://www.kernel.org/doc/html/v6.1/userspace-api/media/rc/lirc-dev-intro.html

## LIRC

No LIRC user space packages are needed. The Panasonic inverter remote control does not transmit simple button presses, therefore the LIRC functionality for mapping IR pulse/space sequences to buttons is of no use. Instead we must ourselves read (and write) the RAW LIRC data provided on the /dev/lirc devices by kernel drivers.

### Configuring the IR modules

The kernel IR modules are overlays that are enabled by editing /boot/firmware/config.txt:

```
# InfraRed
dtoverlay=gpio-ir,gpio_pin=23
#dtoverlay=gpio-ir-tx,gpio_pin=18
dtoverlay=pwm-ir-tx,gpio_pin=18
```

Note that there are two different drivers for transmission:

* `gpio-ir-tx` uses what is known as "bit-banging", where the processor must control the IR output directly. I was not able to get this to work.
* `pwm-ir-tx` uses PWM to control the IR output. This does not require as much of the processor, but it still needs to trigger the edges, and is therefore sensitive to delays. The duration of pulses and spaces is therefore not very precise, and I had problems where transmissions were not reliably accepted by the receiver.

To be able to test both transmission drivers, I chose GPIO pin 18 for output, since it can be used with PWM. In the end, I was only able to get PWM to work.

Note that the receiver output is high at rest and drops to low when it receives IR light. The GPIO pin should therefore be configured with pull-up, which is the default for the gpio-ir module. I chose GPIO pin 23 for the input, becasue it was physically close to pin 18.

### Using a custom Raspberry Pi OS kernel

At the time of writing (February 2024), a patch for `pwm-ir-tx` using hrtimers (high resolution timers) is being merged into the Linux upstream kernel, but is not yet available for the Raspberry Pi. I believe they have been merged into the latest upstream kernel, 6.8, but have not been backported yet. Raspberry Pi OS is still based on 6.1. The Raspberry Pi OS repo has a branch `rpi-6.6.y`, and the patches applied with only a little fuzz. With these patches, the `pwm-ir-tx` kernel module works _much_ better.

Building a custom kernel: https://www.raspberrypi.com/documentation/computers/linux_kernel.html#kernel

Raspberry Pi OS 6.6 branch: https://github.com/raspberrypi/linux/tree/rpi-6.6.y

Patches for the `pwm-ir-tx` driver:  https://lore.kernel.org/linux-pwm/cover.1703003288.git.sean@mess.org/

## Configuring udev

A problem when both the receiver and transmitter modules are enabled is that the kernel will create to devices, /dev/lirc0 and /dev/lirc1, but there is no direct way to know which is which. Therefore it is useful to add udev rules like this:

/etc/udev/rules.d/70-lirc.rules

```
ACTION=="add", SUBSYSTEM=="lirc", DRIVERS=="gpio_ir_recv", SYMLINK+="lirc-rx"
ACTION=="add", SUBSYSTEM=="lirc", DRIVERS=="gpio-ir-tx", SYMLINK+="lirc-tx"
ACTION=="add", SUBSYSTEM=="lirc", DRIVERS=="pwm-ir-tx", SYMLINK+="lirc-tx"
```

These rules will match on the driver name, and create symbolic links that point to the associated device. This allows you to use `/dev/lirc-rx` for reception and `/dev/lirc-tx` for transmission.

## Configuring limits

To allow running a process with a higher priority, the following can be added:

/etc/security/limits.d/pi.conf

```
pi     -       nice    -20
```

This will allow the `pi` user to set process niceness all the way to -20 (the highest priority).
