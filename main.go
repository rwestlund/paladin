/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */
package main

import ()

/* The location of the configuration file to read. */
const config_file = "superv.conf"

func main() {
	var g Global = Global{}
	/* Initialize pointer types. */
	g.RunningChan = make(chan LaunchStatus)
	g.DoneChan = make(chan LaunchStatus)
	g.Procs = make(map[string]*Process)

	/* Read the config file. */
	var config *Config = ParseConfigFile(config_file)
	// Build the global object.
	for _, pc := range config.Process {
		g.Procs[pc.Name] = &Process{Config: pc}
	}

	// Launch only leaf node processes; those that don't depend on any others.
	// The remaining processes will be launched when the events from these
	// starting are received.
	for i := range g.Procs {
		if len(g.Procs[i].Config.SoftDepends) == 0 {
			/* Launch the process in a new goroutine. */
			go LaunchProcess(g.Procs[i].Config, &g)
			g.RunningProcesses++
		}
	}

	for g.RunningProcesses > 0 {
		select {
		/* Listen for events fired when a child starts. */
		case status := <-g.RunningChan:
			HandleRunning(&g, status)

		/* Listen for events fired when a child exits. */
		case status := <-g.DoneChan:
			g.RunningProcesses--
			HandleDone(&g, status)
		}
	}
}
