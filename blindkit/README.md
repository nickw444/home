# Blindkit

Control RAEX Motorised blinds by emulating their 433.92MHz RF protocol.

- [esp-firmware](esp-firmware): WIP: A ESP8266 firmware for sending remote codes to blinds using an RF transmitter. Interface via MQTT.
- [remote-gen](remote-gen): A (golang) tool to generate new remote codes and calculate the apropriate checksum.
- [rf-process](rf-process): A collection of tooling for converting 433MHz signal captures in WAV files to binary data for analysis and processing.

### [Writeup](https://nickwhyte.com/post/2017/reversing-433mhz-raex-motorised-rf-blinds/)

A detailed writing explaining the process I took to reverse engineer the protocol RAEX are using to control the motorised blinds.

### Research Resources, Scripts & Tools.

- [RF Blinds Spreadsheet](https://docs.google.com/spreadsheets/d/1oP6-OY93fNaIKRSyX8hcdRp30glidLyhRHn7Lt-4DNo/edit?usp=sharing) A spreadsheet where I worked on analysing the processed data from the `rf-process` tool, identifying visual patterns in the bits.
- [sketches/receive_manchester.ino](sketches/receive_manchester.ino) Sniff 433MHz transmissions.
- [sketches/transmit.ino](sketches/transmit.ino) An early prototype of transmitting a captured remote code.
- [research/initial-captures.txt](research/initial-captures.txt) Early captures taken via `receive_manchester.ino` sketch.
- [research/initial-captures-analysis.txt](research/initial-captures-analysis.txt) Early capture analysis of `initial-captures.txt`.


#### Links:
- [`YR1526` Remote on FCCID](https://fccid.io/FLS-YR1526)
- [Reversing via dumping transmitter firmware](http://travisgoodspeed.blogspot.co.uk/2010/07/reversing-rf-clicker.html)
- [Somfy blind protocol](https://pushstack.wordpress.com/somfy-rts-protocol/), [Reversing somfy RTS](https://pushstack.wordpress.com/2014/04/14/reversing-somfy-rts/)
- [BlindsT4 in RFXTRX](https://github.com/domoticz/domoticz/blob/master/main/RFXtrx.h#L651)
- [Controlling Blinds.com RF Dooya Motors with Arduino and Vera](https://forum.mysensors.org/topic/7/controlling-blinds-com-rf-dooya-motors-with-arduino-and-vera)
- [433MHz RF Modules](http://www.lydiard.plus.com/hardware_pages/433mhz_modules.htm)
- [ArduinoWeatherOS](https://github.com/robwlakes/ArduinoWeatherOS)
- [A Lesson In Blind Reverse Engineering](http://www.rtl-sdr.com/blind-reverse-engineering-wireless-protocol/), [PDF Writeup](https://github.com/LucaBongiorni/Amateur-SIGINT/blob/master/Amateur%20Signals%20Intelligence.pdf)
- [Reverse Engineer Wireless Temperature / Humidity / Rain Sensors](http://rayshobby.net/reverse-engineer-wireless-temperature-humidity-rain-sensors-part-1/)
