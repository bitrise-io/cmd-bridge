#!/bin/bash

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -v
set -e

run_params=$@

bash "${THIS_SCRIPT_DIR}/build.sh"

echo "Params: ${run_params}"

"${THIS_SCRIPT_DIR}/../bin/osx/cmd-bridge" "${run_params}"
