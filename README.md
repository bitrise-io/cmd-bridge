# cmd-runner-miniserver

Stripped down, minimal server written in [go](https://golang.org/)
which accepts a command line command, executes it,
logs it's output and returns a JSON response
with the exit status.

## Build & run

build: `$ bash _scripts/build.sh`

run: `$ ./bin/osx/cmd-runner-miniserver`

Or in one command: `$ bash _scripts/build.sh && ./bin/osx/cmd-runner-miniserver`

## Usage

Once the server runs you can use it through HTTP messages.

For example:

    curl http://localhost:27473/ping

A simple `echo 'Hello world!'`:

    curl -X POST -d "{\"command\": \"echo 'Hello world!'\"}" http://localhost:27473/cmd

Echo a supplied environment variable:

    curl -X POST -d '{"command":"echo \"Hello: ${T_KEY}!\"","environments":[{"key":"T_KEY","value":"test value, with equal = sign, for test"}]}' http://localhost:27473/cmd

Use the included `_scripts/gen_json.rb` to generate the content (JSON) for cURL:

    curl -X POST -d "$(ruby _scripts/gen_json.rb)" http://localhost:27473/cmd

Run a bash script:

    curl -X POST -d '{"command":"bash /path/to/script.sh"}' http://localhost:27473/cmd

If you specify the script's path through a (environment) variable:

    export SCRIPT_PTH=/path/to/script
    curl -X POST -d "{\"command\":\"bash ${SCRIPT_PTH}\"}" http://localhost:27473/cmd

## TODO

* Should handle Environment Variables (specify it for the command) but **should not** add it to it's own environment, in order to keep a "clean" command host environment.
  * This means that the supported Environment Variables have to be expanded before sending to this server. No `os.Setenv` should be used!
