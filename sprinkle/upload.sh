#!/bin/sh
set -eux

GOOS=linux GOARCH=arm GOARM=5 go build -o sprinkle-homekit-armv6
scp sprinkle-homekit-armv6 sprinkle@sprinkle:/sprinkle/bin/sprinkle-homekit