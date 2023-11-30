
https://github.com/raspberrypi/linux/issues/2993#issuecomment-497420228

https://www.kernel.org/doc/html/v6.1/userspace-api/media/rc/lirc-dev.html

https://www.kernel.org/doc/html/v6.1/userspace-api/media/rc/lirc-dev-intro.html

Note that the receiver output is high at rest and drops to low when it receives IR light. The GPIO pin should therefore be configured with pull-up, which is the default for the gpio-ir module.

