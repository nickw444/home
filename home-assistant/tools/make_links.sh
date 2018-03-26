#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd `dirname $0` && pwd)

COMPONENTS_BASE="./components";
CONFIG_BASE="./config";


main() {
    cd "${SCRIPT_DIR}/../conf"

    rm -rf config

    find "${COMPONENTS_BASE}" -type f -name '*.yaml' | while read -r config_file; do
        local cleaned_file=$(echo $config_file | sed "s@${COMPONENTS_BASE}/@@");
        local component_name=$(echo $(dirname ${cleaned_file}))
        local filename=$(basename "${config_file}")
        local config_type=$(echo "${filename%.*}")

        local link_destination="${CONFIG_BASE}/${config_type}/${component_name}.yaml";
        echo "${component_name} (${config_type}) -> ${link_destination}"

        mkdir -p "${CONFIG_BASE}/${config_type}";
        ln -s "../../${config_file}" "${link_destination}"

    done
}

main "$@"
