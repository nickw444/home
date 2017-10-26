# bridge

The main Homekit bridge. Bridges MQTT topics to logical Homekit accessories. 
Currently only supports:

 - Garage Door (via `esp-fw/garagedoor`)
 - Sonoff Relay (via `esp-fw/sonoff-relay`)
 - Sonoff TH10/TH16 Temp/Humidty Sensor (via `esp-fw/sonoff-th10`)

Future plans including extending this accessory library to:

 - Door (via `esp-fw/door-trigger`)
 - Electric Blinds (via `esp-fw/433-hub`)

## Usage

```
./bridge --help
usage: MQTTBridge [<flags>] <accessCode>

Homekit MQTT Bridge

Flags:
  --help                       Show context-sensitive help (also try --help-long and --help-man).
  --config="bridge.conf.yml"   Provide a configuration file.
  --port=PORT                  Port for Homekit to listen on.
  --mqttBroker="tls://127.0.0.1:8883"
                               MQTT Broker URL
  --mqttUser=MQTTUSER          MQTT Broker User
  --mqttPassword=MQTTPASSWORD  MQTT Password

Args:
  <accessCode>  Homekit Access code to use
```


## Configuration

The bridge accepts a configuration file, describing the MQTT devices that you
wish to bridge.


### Example Config

```yaml
name: MainBridge
manufacturer: nickw
model: homebridge

accessories:
  - model: sonoff-switch
    serial: 5ccf7f9551dc
    name: Test Room

  - model: sonoff-th10
    serial: 60194173177
    name: TempTest

  - model: garagedoor
    serial: 0000
    name: GarageTest

```


## Building

This is a go project, so clone it into your GOPATH using `go get`, then,

```
go build
```

