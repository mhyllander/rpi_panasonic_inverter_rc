# Configuring the Raspberry Pi

## Some reference material

https://github.com/raspberrypi/linux/issues/2993#issuecomment-497420228

https://www.kernel.org/doc/html/v6.1/userspace-api/media/rc/lirc-dev.html

https://www.kernel.org/doc/html/v6.1/userspace-api/media/rc/lirc-dev-intro.html

https://www.mess.org/2020/01/26/Moving-from-lirc-tools-to-rc-core-tooling/

## Using a custom Raspberry Pi OS kernel

At the time of writing (February 2024), a patch for `pwm-ir-tx` using hrtimers (high resolution timers) has been merged into Linux kernel 6.8, but is not yet available for the Raspberry Pi. Raspberry Pi OS is still based on 6.1, on the verge of upgrading to 6.6. The Raspberry Pi OS repo has a branch `rpi-6.6.y`, and the patches applied with only a little fuzz. With these patches, the `pwm-ir-tx` kernel module works _much_ better. _Note: It turns out that the patches work well when applied to 6.6.13, but not so much with 6.6.16 or 6.6.17._

Building a custom kernel: https://www.raspberrypi.com/documentation/computers/linux_kernel.html#kernel

Raspberry Pi OS 6.6 branch: https://github.com/raspberrypi/linux/tree/rpi-6.6.y

Patches for the `pwm-ir-tx` driver:  https://lore.kernel.org/linux-pwm/cover.1703003288.git.sean@mess.org/

## Initial setup

```bash
sudo apt update
sudo apt full-upgrade -y
sudo reboot
sudo rpi-update # currently updates to rpi-6.6.y
sudo reboot
# optional
sudo apt install -y etckeeper
sudo apt install -y unattended-upgrades
```

## Installing custom kernel

```bash
DATE=$(date +'%Y-%m-%d')

sudo cp -a /boot/firmware /boot/firmware.backup.$DATE
sudo cp -r BUILD/boot/firmware /boot/firmware

# Assuming we are patching the current kernel version
VERSION=$(uname -r)
sudo cp -a /lib/modules/$VERSION /lib/modules/$VERSION.backup.$DATE
sudo cp -r BUILD/lib/modules/$VERSION /lib/modules/$VERSION

# add "" to /boot/firmware/config.txt
```

Add `kernel=kernel8-custom.img` to `/boot/firmware/config.txt`.

### HDMI no signal issue

My screen went black after initial boot messages. No login prompt. I found this [trouble-shooting guide](https://pip.raspberrypi.com/categories/685-whitepapers-app-notes/documents/RP-004341-WP/Troubleshooting-KMS-HDMI-output.pdf).

```bash
cat /sys/class/drm/card?-HDMI-A-1/edid
```

did not return an EDID, suggesting that KMS had problems reading the displays EDID. Adding the following to `/boot/firmware/cmdline.txt` solved the problem:

```
video=HDMI-A-1:1280x720@60D
```

## LIRC

No LIRC user space packages are needed. The Panasonic inverter remote control does not transmit simple button presses, therefore the LIRC functionality for mapping IR pulse/space sequences to buttons is of no use. Instead we must ourselves read (and write) the RAW LIRC data provided on the /dev/lirc devices by kernel drivers.

### Configuring the IR modules

The kernel IR modules are overlays that are enabled by editing `/boot/firmware/config.txt`:

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

### Configuring udev

A problem when both the receiver and transmitter modules are enabled is that the kernel will create to devices, /dev/lirc0 and /dev/lirc1, but there is no direct way to know which is which. Therefore it is useful to add udev rules like this:

`/etc/udev/rules.d/70-lirc.rules`

```
ACTION=="add", SUBSYSTEM=="lirc", DRIVERS=="gpio_ir_recv", SYMLINK+="lirc-rx"
ACTION=="add", SUBSYSTEM=="lirc", DRIVERS=="gpio-ir-tx", SYMLINK+="lirc-tx"
ACTION=="add", SUBSYSTEM=="lirc", DRIVERS=="pwm-ir-tx", SYMLINK+="lirc-tx"
```

These rules will match on the driver name, and create symbolic links that point to the associated device. This allows you to use `/dev/lirc-rx` for reception and `/dev/lirc-tx` for transmission.

## Configuring limits

To allow running a process with a higher priority, the following can be added:

`/etc/security/limits.d/pi.conf`

```
pi     -       nice    -20
```

This will allow the `pi` user to set process niceness all the way to -20 (the highest priority).

## Configuring unattended upgrades

`/etc/apt/apt.conf.d/50unattanded-upgrades`

```diff
diff --git a/apt/apt.conf.d/50unattended-upgrades b/apt/apt.conf.d/50unattended-upgrades
index 7fbd3d4..b26d125 100644
--- a/apt/apt.conf.d/50unattended-upgrades
+++ b/apt/apt.conf.d/50unattended-upgrades
@@ -26,11 +26,12 @@ Unattended-Upgrade::Origins-Pattern {
         // archives (e.g. from testing to stable and later oldstable).
         // Software will be the latest available for the named release,
         // but the Debian release itself will not be automatically upgraded.
-//      "origin=Debian,codename=${distro_codename}-updates";
+        "origin=Debian,codename=${distro_codename}-updates";
 //      "origin=Debian,codename=${distro_codename}-proposed-updates";
         "origin=Debian,codename=${distro_codename},label=Debian";
         "origin=Debian,codename=${distro_codename},label=Debian-Security";
         "origin=Debian,codename=${distro_codename}-security,label=Debian-Security";
+        "a=stable,c=main,o=Raspberry Pi Foundation,l=Raspberry Pi Foundation";

         // Archive or Suite based matching:
         // Note that this will silently match a different release after
@@ -45,7 +46,7 @@ Unattended-Upgrade::Origins-Pattern {
 // Python regular expressions, matching packages to exclude from upgrading
 Unattended-Upgrade::Package-Blacklist {
     // The following matches all packages starting with linux-
-//  "linux-";
+    "linux-";

     // Use $ to explicitely define the end of a package name. Without
     // the $, "libc6" would match all of them.
@@ -112,7 +113,7 @@ Unattended-Upgrade::Package-Blacklist {

 // Automatically reboot *WITHOUT CONFIRMATION* if
 //  the file /var/run/reboot-required is found after the upgrade
-//Unattended-Upgrade::Automatic-Reboot "false";
+Unattended-Upgrade::Automatic-Reboot "true";

 // Automatically reboot even if there are users currently logged in
 // when Unattended-Upgrade::Automatic-Reboot is set to true
```
