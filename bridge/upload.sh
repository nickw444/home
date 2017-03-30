#!/bin/sh
set -eux

GOOS=linux GOARCH=arm GOARM=5 go build -o homekit-bridge-armv6
scp homekit-bridge-armv6 sprinkle@sprinkle:/sprinkle/bin/homebridge2
