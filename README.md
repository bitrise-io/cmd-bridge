# cmd-runner-miniserver

Stripped down, minimal server written in [go](https://golang.org/)
which accepts a command line command, executes it,
logs it's output and returns a JSON response
with the exit status.

## Build & run

build: `$ bash _scripts/build.sh`

run: `$ ./bin/osx/cmd-runner-miniserver`

Or in one (bash) command: `$ bash _scripts/build.sh && ./bin/osx/cmd-runner-miniserver`

## Usage

Once the server runs you can use it through HTTP messages.

For example:

    curl http://localhost:27473/ping

A simple `echo 'Hello world!'`:

    curl -X POST -d "{\"command\": \"echo 'Hello world!'\"}" http://localhost:27473/cmd
