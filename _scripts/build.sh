#!/bin/bash

set -e

THIS_SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "${THIS_SCRIPT_DIR}/.."

set +e

echo "-> Testing"
go test ./...
if [[ $? -ne 0 ]]; then
	echo " [!] Tests Failed"
	exit 1
fi

echo "-> Building"
go build
if [[ $? -ne 0 ]]; then
	echo " [!] Build Failed!"
	exit 1
fi

echo "-> Moving binary"
mkdir -p ./bin/osx
mv ./cmd-bridge ./bin/osx/
if [[ $? -ne 0 ]]; then
	echo " [!] Binary move Failed!"
	ls -alh
	exit 1
fi

echo " (i) DONE - OK"
