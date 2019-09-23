#!/bin/sh
set -eu

CONFIG_PATH=$(dirname "$0")/../

main() {
    local config_dir="${CONFIG_PATH}/$1";

    cd "${config_dir}";

    HA_VERSION=$(<.HA_VERSION)
    mv secrets.example.yaml secrets.yaml

    # pip3 install homeassistant==${HA_VERSION}
    hass --script check_config -c . -f
}

main "$@"
