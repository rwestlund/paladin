/*
 * Copyright (c) 2016-2017, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */

package main

import (
	"log"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

// The location of the configuration file to read. Compile with .e.g.
// -ldflags="-X main.localbase=/usr/local" if the config file will not be under
// /etc/.

var localbase = ""
var configFile = localbase + "/etc/paladin.conf"

func main() {
	var g global = global{}
	// Initialize pointer types.
	g.RunningChan = make(chan launchStatus)
	g.DoneChan = make(chan launchStatus)
	g.Procs = make(map[string]*process)

	// Read the config file.
	var config *configOptions = parseConfigFile(configFile)

	// Change logger output destination, if necessary.
	if config.LogFile != "" {
		logfile, err := os.OpenFile(config.LogFile,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatal("Failed to open log file ", config.LogFile, "\n", err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	// Build the global object.
	for _, pc := range config.Process {
		g.Procs[pc.Name] = &process{Config: pc}
	}

	// Listen for a kill signal. When one comes, kill all running children
	// before exiting.
	var sigs = make(chan os.Signal)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT)
	go func() {
		var sig = <-sigs
		log.Println("Signal", sig, " received, termiating children...")
		for i := range g.Procs {
			if g.Procs[i].Running {
				log.Println("killing", g.Procs[i].Config.Name)
				unix.Kill(g.Procs[i].Status.Pid, unix.SIGTERM)
			}
		}
		log.Println("done")
		os.Exit(0)
	}()

	// Launch only leaf node processes; those that don't depend on any others.
	// The remaining processes will be launched when the events from these
	// starting are received.
	for i := range g.Procs {
		if len(g.Procs[i].Config.SoftDepends) == 0 {
			// Launch the process in a new goroutine.
			go launchProcess(g.Procs[i].Config, &g)
			g.RunningProcesses++
		}
	}
	for g.RunningProcesses > 0 {
		select {
		// Listen for events fired when a child starts.
		case status := <-g.RunningChan:
			handleRunning(&g, status)

		// Listen for events fired when a child exits.
		case status := <-g.DoneChan:
			g.RunningProcesses--
			handleDone(&g, status)
		}
	}
}
