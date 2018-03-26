#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd `dirname $0` && pwd)

main() {
    cd "${SCRIPT_DIR}"
    pipenv run hass --script check_config -c "${SCRIPT_DIR}/../conf"
}

main "$@"

