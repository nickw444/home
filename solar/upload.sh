#!/bin/sh
set -eux

GOOS=linux GOARCH=arm GOARM=5 go build -o solar-homekit-armv6
scp solar-homekit-armv6 sprinkle@sprinkle:/sprinkle/bin/solar-homekit