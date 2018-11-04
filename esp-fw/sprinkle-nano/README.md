# sprinkle-nano

An firmware targeting an Arduino Nano (or similar), connecting to the network using a ENC28J60 ethernet interface. It's interface occurs via an MQTT server.

## Interface

### Topics:
- `sprinkle/status`: retained online/offline status of device
- `sprinkle/n`: retained state topic for output n)
- `sprinkle/n/set`: set state topic for output n. See payload below

### Set Payload

Payload will be in the following formats:

- `ON` - turn on indefinitely
- `OFF` - turn off
- `ON:12320` - turn on for `12320` seconds


## Wiring W5100

- `CS` (`NSS`) -> `D10`
- `MOSI` (`MO`) -> `D11`
- `MISO` (`MI`) -> `D12`
- `SCK` (`SOK`) -> `D13`
- `VCC` -> `5v`
- `GND` -> `GND`
