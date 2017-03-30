# esp-fw

A collection of custom ESP and Arduino firmware used for these automation 
projects.

Most firmware here targets [ITEAD sonoff](https://www.itead.cc/sonoff-wifi-wireless-switch.html) products OR [ESP8266](https://www.sparkfun.com/products/13678) chips.

## Firmware

- [sonoff-relay](sonoff-relay/): Convert an [ITEAD sonoff wireless smart switch](https://www.itead.cc/sonoff-wifi-wireless-switch.html) into an MQTT connected one.
- [sonoff-th10](sonoff-th10): Convert an [ITEAD TH10/TH16](https://www.itead.cc/smart-home/sonoff-th.html) temp/humidty sensor into an MQTT connected one.


## Libraries

- [SimpleMQTT](lib/SimpleMQTT): A library to generalise connection, subscription and publishes from an ESP/Arduino to an MQTT broker.


## Sonoff/ESP/FW Resources

- https://www.hackster.io/idreams/getting-started-with-sonoff-rf-98a724
- https://www.itead.cc/blog/sonoff-esp8266-reprogramming-part-1-control-mains-from-anywhere
- http://www.esp8266.com/wiki/doku.php?id=esp8266-module-family
- http://www.forward.com.au/pfod/ESP8266/GPIOpins/ESP8266_01_pin_magic.html
- http://www.allaboutcircuits.com/projects/breadboard-and-program-an-esp-01-circuit-with-the-arduino-ide/
