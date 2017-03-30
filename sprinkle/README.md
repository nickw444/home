# sprinkle

A standalone accessory to control Sprinkler relays connected via RPi GPIO.

## Usage

```
./sprinkle --help
usage: sprinkle [<flags>] <command> [<args> ...]

Homekit Sprinkler Control

Flags:
  --help                        Show context-sensitive help (also try --help-long and --help-man).
  --config="sprinkle.conf.yml"  Provide a configuration file.

Commands:
  help [<command>...]
    Show help.

  run* [<flags>] <accessCode> [<port>]
    Run the HAP Server.

  sample-config <num-circuits>
    Generate a sample config file
```

```
./sprinkle run --help
usage: sprinkle run [<flags>] <accessCode> [<port>]

Run the HAP Server.

Flags:
  --help                        Show context-sensitive help (also try --help-long and --help-man).
  --config="sprinkle.conf.yml"  Provide a configuration file.
  --without-gpio                Disable GPIO. Useful for devlepment on a platform without GPIO

Args:
  <accessCode>  Homekit Access code to use
  [<port>]      Port for homekit to listen on.
```


## Configuration

A configuration file is passed to specify the connected relays. You can 
generate a configuration file with `./sprinkle sample-config <number of relays>`

### Example Configuration

```yaml
manufacturer: My Name
bridge:
  name: MyBridge
  serial: bridge-00001
circuits:
- name: Circuit 1
  bcmPort: 2
  maxDuration: 0
  serial: circuit-00000
- name: Circuit 2
  bcmPort: 3
  maxDuration: 0
  serial: circuit-00001
- name: Circuit 3
  bcmPort: 4
  maxDuration: 0
  serial: circuit-00002
```

Keys for circuits are:

 - `bcmPort`: The bcm port number that the relay is connected to
 - `maxDuration`: Time in minutes to automatically switch off after. If 0 is 
                  provided, the relay will never be automatically switched off.
 - `serial`: Serial to show in Homekit.
 - `name`: Name to show in Homekit.


## Building

This is a go project, so clone it into your GOPATH using `go get`, then,

```
go build
```