# Superv

A simple process supervisor written in Go.

## Description

Superv (pronounced 'super vee') is a simple way to launch and maintain
services. It provides the following features:
- Processes to be launched are specified in a TOML config file.
- Processes that fail can be restarted.
- Each process can log to different files for stdout and stderr.

## Installation

- Clone this repo in your `$GOPATH`
- Edit `superv.conf`
- Run `go run main.go`

## License

This code is under the BSD-2-Clause license.  See the LICENSE file for the full
text.

## TODO

There are still many features that should be implemented before this is useful
enough for a production system, including:
- Process dependencies
- Proper config file location
- Main process logging
- PID files
