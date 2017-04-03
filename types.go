/*
 * Copyright (c) 2016-2017, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This file contains struct definitions.
 */

package main

import (
	"time"
)

// The status object returned from launching a child in a goroutine.
type launchStatus struct {
	Name string
	Pid  int
	// If the process failed for any reason, that reason is here.
	Err error
	// The duration for which the process ran.
	Duration time.Duration
}

type process struct {
	// The process config object from the config file.
	Config processConfig
	// The most recent launch status.
	Status launchStatus
	// Whether it's currently running or not.
	Running bool
}

type global struct {
	// A dictionary of all processes.
	Procs map[string]*process
	// Count the children we have active so we know when to exit.
	RunningProcesses int
	// Listen on this channel to know when a program has started.
	RunningChan chan launchStatus
	// Listen on this channel to know when a program is done.
	DoneChan chan launchStatus
}

// A process definition, as read directly from the config file.
type processConfig struct {
	// A human-readable tag for process.
	Name string
	// The path to the actual executable to run.
	Path string
	// A string with arguments to be passed to the process.
	Args string
	// The cwd of the process. Defaults to /var/empty.
	Cwd string
	// Filenames for writing stdout and stderr. Defaults to /dev/null.
	Stdout string
	Stderr string
	// The number of milliseconds to wait before restarting a process.
	RestartDelay uint64 `toml:"restart_delay"`
	// Whether to disable restarting on failure.
	IgnoreFailure bool `toml:"ignore_failure"`
	// If a process exits within this many milliseconds, don't restart it. A
	// value of 0 disables this check.
	MinRuntime  int      `toml:"min_runtime"`
	SoftDepends []string `toml:"soft_depends"`
	User        string
	Group       string
}

// The config file definition.
type configOptions struct {
	// This must be named after the [[process]] block in the config file.
	Process []processConfig
	// Where to send paladin's logging output. Defaults to stderr.
	LogFile string `toml:"log_file"`
}
