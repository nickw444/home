#!/bin/sh
set -eu

CONFIG_PATH=$(dirname "$0")/../

main() {
    local config_dir="${CONFIG_PATH}/$1";

    cd "${config_dir}";

    HA_VERSION=$(cat .HA_VERSION);
    cp secrets.example.yaml secrets.yaml;
    echo "Installing homeassistant==${HA_VERSION}";
    pip3 install hass-deps "homeassistant==${HA_VERSION}";
    if [[ -f "hass-deps.yaml" ]]; then
      hass-deps install
    fi

    # Workaround to "not a directory @ data['whitelist_external_dirs'][0]" failure in CI
    sudo mkdir -p /config;

    hass --script check_config -c . -f;
}

main "$@"
