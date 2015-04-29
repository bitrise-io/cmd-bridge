# cmd-bridge

Stripped down, minimal server written in [go](https://golang.org/)
which accepts a command line command, executes it,
logs it's output and returns a JSON response
with the exit status.


## Build & run (for development)

Don't forget to `$ go get ./...`

build: `$ bash _scripts/build.sh`

run in server mode: `$ ./bin/osx/cmd-bridge`

Or in one command: `$ bash _scripts/build.sh && ./bin/osx/cmd-bridge`

Or with the included *build-and-run* script: `$ bash _scripts/build_and_run.sh`

run in non-server mode: `$ ./bin/osx/cmd-bridge -help`

Or in one command: `$ bash _scripts/build.sh && ./bin/osx/cmd-bridge -help`

Or with the included *build-and-run* script: `$ bash _scripts/build_and_run.sh -help`


## Usage

### Server mode

Once the server runs you can use it through HTTP messages.

For example:

    curl http://localhost:27473/ping

A simple `echo 'Hello world!'`:

    curl -X POST -d "{\"command\": \"echo 'Hello world'\"}" http://localhost:27473/cmd

Echo a supplied environment variable:

    curl -X POST -d '{"command":"echo \"Hello: ${T_KEY}!\"","environments":[{"key":"T_KEY","value":"test value, with equal = sign, for test"}]}' http://localhost:27473/cmd

Use the included `_scripts/gen_json.rb` to generate the content (JSON) for cURL:

    curl -X POST -d "$(ruby _scripts/gen_json.rb)" http://localhost:27473/cmd

Run a bash script:

    curl -X POST -d '{"command":"bash /path/to/script.sh"}' http://localhost:27473/cmd

If you specify the script's path through a (environment) variable:

    export SCRIPT_PTH=/path/to/script
    curl -X POST -d "{\"command\":\"bash ${SCRIPT_PTH}\"}" http://localhost:27473/cmd


### Non-server mode

*Running commands requires a running cmd-bridge in server mode.*

Print help: `$ bash _scripts/build_and_run.sh -help`

Run a bash script: `$ bash _scripts/build_and_run.sh -do 'bash /path/to/script'`

**You can also pass environments** for your command. Environment variables
available for the non-server mode process will be sent to the server
process if you prefix the environment key with `_CMDENV__`.

For example to send an environment variable with key `MY_TEST_ENV` to
the server process you should store the value in an environment
with key `_CMDENV__MY_TEST_ENV` so that it's available for the non-server
mode process. The non-server process will collect all of these environments,
remove the special prefix and send it to the server mode process.

An example:

    $ export _CMDENV__ECHO_THIS_ENV='this environment variable will be available for the server mode process, as ECHO_THIS_ENV'
    $ bash _scripts/build_and_run.sh -do='echo "ECHO_THIS_ENV: ${ECHO_THIS_ENV}"'


## Release a new version

1. Bump version in `version.go`
2. `$ bash _scripts/build.sh`
3. Commit
4. Tag the commit with the same version defined in *version.go*
