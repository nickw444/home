#!/bin/bash
set -euo pipefail

CONFIG_DIR=$(cd "$(dirname "$0")"/.. && pwd)

for arg in sd{a..m}; do
    /usr/sbin/smartctl --info --all --json  \
        --nocheck standby /dev/$arg > "${CONFIG_DIR}/smartctl/$arg.json";
done
