#!/bin/bash

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

set -v
set -e

bash "${THIS_SCRIPT_DIR}/build.sh"

echo "Params: $@"

"${THIS_SCRIPT_DIR}/../bin/osx/cmd-bridge" "$@"
