# Superv

A simple process supervisor written in Go.

## Description

Superv (pronounced 'super vee') is a simple way to launch and maintain
services. It provides the following features:
- Simple TOML configuration.
- Automatically restart failed processes.
- Log each process to a different file, even separating stdout and stderr.
- Specify process dependencies, so things start in the right order.

## Installation

- Clone this repo in your `$GOPATH`
- Edit `superv.conf`
- Run `go run main.go`

## License

This code is under the BSD-2-Clause license.  See the LICENSE file for the full
text.

## Configuration

The config file uses TOML format. It consists of general options, and several
`[[process]]` blocks that define each process that Superv is responsible for.

### General Options

The following general options do not go inside a TOML block:

|Name      | Required | Description
|----------|---|-----------------------------------------------------
|`log_file`| N | File path for main output. An empty string means stderr.

### Process-specific Options

The following options are per-process, and go in a `[[process]]` block:

|Name            | Required | Description
|----------------|---|-------------------------------------------------------
|`name`          | Y | Used to identify the process.  Must be unique.
|`path`          | Y | The full path to the program to be run.
|`args`          | N | A single string with all arguments for the process.
|`stdout`        | N | The file path to use for logging stdout.
|`stderr`        | N | The file path to use for logging stderr.
|`restart_delay` | N | Milliseconds to wait before restarting.
|`ignore_failure`| N | Boolean. Set to `true` to disable restarting on failure.
|`min_runtime`   | N | Don't restart if it fails in fewer than this many milliseconds.
|`soft_depends`  | N | List of processes that must be started before this one.

### Example Configuration

```
log_file = "/var/log/superv.log"

[[process]]
name = "sleep3"
path = "/bin/sleep"
args = "3"
soft_depends = [ "ls" ]

[[process]]
name = "ls"
path = "/bin/ls"
args = "-l -a"
stdout = "/tmp/ls-output"

```

## TODO

- Proper config file location
- PID files
- Support programs that fork
- Specify a cwd
- Specify user/group
