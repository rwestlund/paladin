/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */
package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

/* Take a process, an id, and notification channel. Launch the process, wait
 * for it to exit, and signal completion by returning a Status struct on the
 * completion channel, which includes a possible error.
 */
func LaunchProcess(pc ProcessConfig, g *Global) {
	log.Println("Process", pc.Name, "\tlaunching")
	/* Convert p.args to a slice, so the process gets separate arguments. */
	var cmd = exec.Command(pc.Path, squeeze(strings.Split(pc.Args, " "))...)

	/* If there's an output file for stdout specified, use it. */
	if pc.Stdout != "" {
		stdout_file, err := os.OpenFile(pc.Stdout,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			log.Println("Failed to open log file", pc.Stdout, "\n", err)
			g.DoneChan <- LaunchStatus{Name: pc.Name, Err: err}
			return
		}
		defer stdout_file.Close()
		cmd.Stdout = stdout_file
		/* If not, use /dev/null. */
	} else {
		cmd.Stdout = nil
	}

	/* If there's an output file for stderr specified, use it. */
	if pc.Stderr != "" {
		stderr_file, err := os.OpenFile(pc.Stderr,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			log.Println("Failed to open log file", pc.Stderr, "\n", err)
			g.DoneChan <- LaunchStatus{Name: pc.Name, Err: err}
			return
		}
		defer stderr_file.Close()
		cmd.Stderr = stderr_file
		/* If not, use /dev/null. */
	} else {
		cmd.Stderr = nil
	}

	/* Fire off the chiled process, then wait for it to complete. */
	var start_time = time.Now()
	var err = cmd.Start()
	if err != nil {
		log.Println("Process", pc.Name, "\tfailed to start")
		g.DoneChan <- LaunchStatus{Name: pc.Name, Err: err}
		return
	}

	/* Signal that the process is running. */
	g.RunningChan <- LaunchStatus{Name: pc.Name}

	err = cmd.Wait()
	var duration = time.Since(start_time)
	if err != nil {
		log.Println("Process", pc.Name, "\tfailed to run")
		g.DoneChan <- LaunchStatus{Name: pc.Name, Err: err, Duration: duration}
		return
	}
	/* Signal completion. */
	g.DoneChan <- LaunchStatus{Name: pc.Name, Err: nil, Duration: duration}
}
