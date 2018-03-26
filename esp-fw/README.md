# esp-fw

A collection of custom ESP and Arduino firmware used for these automation projects.

Firmware here targets [ESP8266](https://www.sparkfun.com/products/13678) chips.

## Firmware

- esp-blindkit: A firmware designed to expose an API via MQTT to transmit RF 433MHz codes to RAEX Blinds. It could be extended to transmit to other brands and models.
- esp-garage: A firmware designed to open/close a garage door and detect its state and publish results using MQTT.

In the future, I may adopt using a library such as [esp-home-lib](https://github.com/OttoWinter/esphomelib), however it is not quite mature enough for mainline adoption. I will re-evaluate this decision in the future.

## iTEAD / Sonoff Firmware

In the past, I had written custom firmware for sonoff devices, however since their popularity has skyrocked over the last year, there are now a vast array of open source firmware available for these devices. Most of my requirements are solved by using [ESPurna](https://github.com/xoseperez/espurna). Alternatively, you could give [Sonoff-Tasmota](https://github.com/arendst/Sonoff-Tasmota) a try!

## Sonoff/ESP/FW Resources

- https://www.hackster.io/idreams/getting-started-with-sonoff-rf-98a724
- https://www.itead.cc/blog/sonoff-esp8266-reprogramming-part-1-control-mains-from-anywhere
- http://www.esp8266.com/wiki/doku.php?id=esp8266-module-family
- http://www.forward.com.au/pfod/ESP8266/GPIOpins/ESP8266_01_pin_magic.html
- http://www.allaboutcircuits.com/projects/breadboard-and-program-an-esp-01-circuit-with-the-arduino-ide/
