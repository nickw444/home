#!/bin/sh
set -eu

CONFIG_PATH=$(dirname "$0")/../

main() {
    local config_dir="${CONFIG_PATH}/$1";

    cd "${config_dir}";

    HA_VERSION=$(cat .HA_VERSION)
    cp secrets.example.yaml secrets.yaml
    echo "Installing homeassistant==${HA_VERSION}";
    pip3 install "homeassistant==${HA_VERSION}"

    # Workaround for whitelist_external_dirs in 313a config.
    mkdir -p /tmp/homekit
    hass --script check_config -c . -f
}

main "$@"
