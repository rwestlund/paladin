/*
 * Copyright (c) 2016-2017, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This file contains the process management logic.
 */

package main

import (
	"log"
	"time"
)

// Called whenever a child launches.
func handleRunning(
	procs map[string]*process,
	runningChan, doneChan chan launchStatus,
	status launchStatus,
	runningProcesses *int,
) {
	log.Println("Process", status.Name, "\trunning as\t", status.Pid)
	var proc *process = procs[status.Name]
	proc.Running = true

	// Check to see whether anything was waiting on this.
	for i := range procs {
		if contains(procs[i].Config.SoftDepends, proc.Config.Name) {
			log.Println("Process", procs[i].Config.Name,
				"\twas waiting for\t", status.Name)
			var ready = true
			for _, v := range procs[i].Config.SoftDepends {
				if procs[v].Running == false {
					ready = false
				}
			}
			if ready {
				go launchProcess(procs[i].Config, runningChan, doneChan)
				*runningProcesses++
			}
		}
	}
}

// Called whenever a child exits. Take appropriate action, such as restarting
// (if noRestart is not true).
func handleDone(
	procs map[string]*process,
	runningChan, doneChan chan launchStatus,
	status launchStatus,
	runningProcesses *int,
	noRestart bool,
) {
	var proc *process = procs[status.Name]

	// If there was an error and we should try to start it again.
	if status.Err != nil && proc.Config.IgnoreFailure == false {
		log.Println("Process", proc.Config.Name, "\tfailed after",
			status.Duration.String())
		// Give up if it failed too quickly.
		if proc.Config.MinRuntime != 0 &&
			time.Duration(proc.Config.MinRuntime)*time.Millisecond > status.Duration {
			log.Println("Process", proc.Config.Name, "\tfailed too quickly. Giving up.")
		} else {
			// If it didn't fail too quickly, continue with restart. Wait the
			// required time before restarting.
			time.Sleep(time.Duration(proc.Config.RestartDelay) * time.Millisecond)

			// Actually restart it.
			go launchProcess(proc.Config, runningChan, doneChan)
			*runningProcesses++
		}
	} else {
		// If the process completed successfully or we don't care.
		log.Println("Process", proc.Config.Name, "\tfinished after",
			status.Duration.String())
	}
}
