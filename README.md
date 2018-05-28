# Paladin

A simple process supervisor written in Go.

## Description

Paladin is a simple way to launch and maintain services. It provides the
following features:
- Simple TOML configuration.
- Automatically restart failed processes.
- Log each process to a different file, even separating stdout and stderr.
- Specify process dependencies, so things start in the right order.
- Launch processes under different users and groups.

## Installation

- Clone this repo in your `$GOPATH`.
- Copy `etc/paladin.conf.sample` to `/etc/paladin.conf` and set your values.
- Run `go build`.
- Put `paldin` in your path.

Compile with something like `-ldflags="-X main.localbase=/usr/local"` on
systems where the config file will not live under `/etc`, such as FreeBSD.

## License

This code is under the BSD-2-Clause license.  See the LICENSE file for the full
text.

## Configuration

The config file uses TOML format. It consists of general options, and several
`[[process]]` blocks that define each process that paladin is responsible for.

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
|`args`          | N | An array of string arguments for the process.
|`cwd`           | N | The current working directory for the process.
|`stdout`        | N | The file path for logging stdout.
|`stderr`        | N | The file path for logging stderr. Follows stdout if unset.
|`user`          | N | Run the process as this user.
|`group`         | N | Run the process as this group.
|`restart_delay` | N | Milliseconds to wait before restarting.
|`ignore_failure`| N | Boolean. Set to `true` to disable restarting on failure.
|`min_runtime`   | N | Don't restart if it fails in fewer than this many milliseconds.
|`soft_depends`  | N | List of processes that must be started before this one.

### Example Configuration

```
log_file = "/var/log/paladin.log"

[[process]]
name = "my-program"
path = "/path/to/my-program"
args = []
cwd = "/path/to"
restart_delay = 1000
min_runtime = 100
stdout = "/tmp/my-program-stdout"
user = "myuser"
group = "mygroup"

[[process]]
name = "my-other-program"
path = "/path/to/my-other-program"
args = ["-a", "-d"]
min_runtime = 100
soft_depends = [ "my-program" ]
restart_delay = 1000
```

### Command-line Options

| Name | Required | Description
|------|----------|------------
|`-f`  | N | The config file to use. Defaults to /etc/paladin.conf (unless localbase is set).

