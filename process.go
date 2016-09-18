/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This file contains the process management logic.
 */
package main

import (
	"log"
	"time"
)

func HandleRunning(g *Global, status LaunchStatus) {
	log.Println("Process", status.Name, "\trunning as\t", status.Pid)

	var proc *Process = g.Procs[status.Name]
	proc.Running = true

	// Check to see whether anything was waiting on this.
	for i := range g.Procs {
		if contains(g.Procs[i].Config.SoftDepends, proc.Config.Name) {
			log.Println("Process", g.Procs[i].Config.Name,
				"\twas waiting for\t", status.Name)
			var ready bool = true
			for _, v := range g.Procs[i].Config.SoftDepends {
				if g.Procs[v].Running == false {
					ready = false
				}
			}
			if ready {
				go LaunchProcess(g.Procs[i].Config, g)
				g.RunningProcesses++
			}
		}
	}
}

/* Called whenever a child exits. Take appropriate action, such as restarting.
 */
func HandleDone(g *Global, status LaunchStatus) {
	var proc *Process = g.Procs[status.Name]

	/* If there was an error and we should try to start it again. */
	if status.Err != nil && proc.Config.IgnoreFailure == false {
		log.Println("Process", proc.Config.Name, "\tfailed after",
			status.Duration.String())
		/* Give up if it failed too quickly. */
		if proc.Config.MinRuntime != 0 &&
			time.Duration(proc.Config.MinRuntime)*time.Millisecond >
				status.Duration {
			log.Println("Process", proc.Config.Name,
				"\tfailed too quickly. Giving up.")
			/* If it didn't fail too quickly, continue with restart. */
		} else {
			/* Wait the required time before restarting. */
			time.Sleep(time.Duration(proc.Config.RestartDelay) * time.Millisecond)

			/* Actually restart it. */
			go LaunchProcess(proc.Config, g)
			g.RunningProcesses++
		}

		/* If the process completed successfully or we don't care. */
	} else {
		log.Println("Process", proc.Config.Name, "\tfinished after",
			status.Duration.String())
	}
}
