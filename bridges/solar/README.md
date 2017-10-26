# solar

A standalone accessory for transporting Efergy API data into Homekit. 

This project is far from configurable, it has a lot of hard coded logic
around calculating load values based on a complex physical metering setup 
at the mains power board.

See `PowerService.Update`


## Usage
```
./solar --help
usage: solar [<flags>] <accessToken> <accessCode> [<port>]

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Args:
  <accessToken>  Access token used to connect to efergy API
  <accessCode>   Homekit Access code to use
  [<port>]       Port for homekit to listen on
```

## Building

This is a go project, so clone it into your GOPATH using `go get`, then,

```
go build
```
