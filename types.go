/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This file contains struct definitions.
 */
package main

import (
	"time"
)

/* The status object returned from launching a child in a goroutine. */
type LaunchStatus struct {
	Name string
	Pid  int
	/* If the process failed for any reason, that reason is here. */
	Err error
	/* The duration for which the process ran. */
	Duration time.Duration
}

type Process struct {
	// The process config object from the config file.
	Config ProcessConfig
	// The most recent launch status.
	Status LaunchStatus
	// Whether it's currently running or not.
	Running bool
}

type Global struct {
	/* A dictionary of all processes. */
	Procs map[string]*Process
	/* Count the total number of children we have active so we know when to
	 * exit. */
	RunningProcesses int
	/* Listen on this channel to know when a program has started. */
	RunningChan chan LaunchStatus
	/* Listen on this channel to know when a program is done. */
	DoneChan chan LaunchStatus
}

/* A process definition, as read directly from the config file. */
type ProcessConfig struct {
	/* A human-readable tag for process. */
	Name string
	/* The path to the actual executable to run. */
	Path string
	/* A string with arguments to be passed to the process. */
	Args string
	/* Filenames for writing stdout and stderr. Defaults to /dev/null. */
	Stdout string
	Stderr string
	/* The number of milliseconds to wait before restarting a process. */
	RestartDelay uint64 `tome:"restart_delay"`
	/* Whether to disable restarting on failure. */
	IgnoreFailure bool `toml:"ignore_failure"`
	/* If a process exits within this many milliseconds, don't restart it. A
	 * value of 0 disables this check.
	 */
	MinRuntime  int      `toml:"min_runtime"`
	SoftDepends []string `toml:"soft_depends"`
	User        string
	Group       string
}

/* The config file definition. Currently, there are only [[process]] blocks,
 * but global options may be added in the future.
 */
type Config struct {
	/* This must be named after the [[process]] block in the config file. */
	Process []ProcessConfig
	// Where to send superv's logging output. Defaults to stderr.
	LogFile string `toml:"log_file"`
}
