# Cloned from https://github.com/brutella/hklifx
This repo was cloned from https://github.com/brutella/hklifx as some changes
were required around vendoring of dependencies.

# hklifx

This is a HomeKit bridge for [LIFX](http://www.lifx.com) light bulbs using [HomeControl](https://github.com/brutella/hc) and ~~[lifx](https://github.com/wolfeidau/lifx)~~ [golifx](https://github.com/pdf/golifx).

LIFX light bulbs are automatically discovered and published as HomeKit accessories on your local network.
After pairing the light bulbs with HomeKit using any iOS HomeKit app (e.g. [Home][home]), you can 

- use Siri to control your lights using voice command – *Hey Siri turn on the bedroom lights*
- use HomeKit triggers to automate your lights – turn on the lights every day at 7:00 AM
- remotely access your lights using HomeKit Remote Access (HomeKit uses strong end-to-end encryption)

# Getting Started

1. [Install Go](http://golang.org/doc/install)
2. [Setup Go workspace](http://golang.org/doc/code.html#Organization)
3. Install

        cd $GOPATH/src
        
        # Clone project
        git clone https://github.com/brutella/hklifx && cd hklifx
        
        # Fetch hklifxd go dependencies
        go get
4. Run

        go run hklifxd.go -pin 00102003 -v

5. Pairing: The official [LIFX app](http://www.lifx.com/pages/go) for iOS or Android is required to initially setup the light bulbs. After that you can use the `hklifxd` daemon to control your lights via HomeKit by using [Home][home] or any other HomeKit-compatible app.

[home]: http://selfcoded.com/home/

**Command Line Arguments**

Required

- `-pin [8-digits]` must be entered on iOS to pair with the light bulb(s)

Optional

- `-transition-duration [seconds]` sets the transition speed
- `-v` for verbose log output

# Contributors

- Mark Wolfe ([wolfeidau](https://github.com/wolfeidau))
- Pieter Maene ([Pmaene](https://github.com/Pmaene))
- Peter Fern ([pdf](https://github.com/pdf))

# Contact

Matthias Hochgatterer

Github: [https://github.com/brutella](https://github.com/brutella)

Twitter: [https://twitter.com/brutella](https://twitter.com/brutella)

# License

hklifx is available under a non-commercial license. See the LICENSE file for more info.
