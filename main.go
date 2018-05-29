/*
 * Copyright (c) 2016-2017, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */

package main

import (
	"flag"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"os/signal"
	"text/tabwriter"
)

// The location of the configuration file to read. Compile with .e.g.
// -ldflags="-X main.localbase=/usr/local" if the config file will not be under
// /etc/.

func printYN(b bool) string {
	if b {
		return "Yes"
	} else {
		return "No"
	}
}

var localbase = ""

func main() {
	// A dictionary of all processes.
	procs := make(map[string]*process)
	// Count the children we have so we know when to exit.
	runningProcesses := 0
	// Listen on this channel to know when a program has started.
	runningChan := make(chan launchStatus)
	// Listen on this channel to know when a program is done.
	doneChan := make(chan launchStatus)
	// This writer will be used to print info in response to SIGINFO.
	writer := tabwriter.NewWriter(os.Stderr, 0, 8, 0, '\t', 0)

	// Read the config file.
	var configFile = flag.String("f", localbase+"/etc/paladin.conf",
		"The configuration file to read from.")
	flag.Parse()
	var config *configOptions = parseConfigFile(*configFile)

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

	// Build the process dictionary.
	for _, pc := range config.Process {
		procs[pc.Name] = &process{Config: pc}
	}

	// Listen for signals and respond appropriately.
	var sigs = make(chan os.Signal)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGINFO)
	go func() {
		for true {
			var sig = <-sigs
			switch sig {
			// In the case of a kill signal, terminate all
			// children before exiting.
			case unix.SIGTERM, unix.SIGINT:
				log.Println("Signal", sig, " received, termiating children...")
				for i := range procs {
					if procs[i].Running {
						log.Println("killing", procs[i].Config.Name)
						unix.Kill(procs[i].Status.Pid, unix.SIGTERM)
					}
				}
				log.Println("done")
				os.Exit(0)
			case unix.SIGINFO:
				fmt.Println("The following children are running:")
				fmt.Fprintln(writer, "Process\tRunning\tPath\tArgs\tUser\tGroup")
				fmt.Fprintln(writer, "-------\t-------\t----\t----\t----\t-----")
				for i := range procs {
					fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", procs[i].Config.Name,
						printYN(procs[i].Running), procs[i].Config.Path,
						procs[i].Config.Args, procs[i].Config.User, procs[i].Config.Group)
					writer.Flush()
				}
			}
		}
	}()

	// Launch only leaf node processes; those that don't depend on any others.
	// The remaining processes will be launched when the events from these
	// starting are received.
	for i := range procs {
		if len(procs[i].Config.SoftDepends) == 0 {
			// Launch the process in a new goroutine.
			go launchProcess(procs[i].Config, runningChan, doneChan)
			runningProcesses++
		}
	}
	for runningProcesses > 0 {
		select {
		// Listen for events fired when a child starts.
		case status := <-runningChan:
			procs[status.Name].Status = status
			handleRunning(procs, runningChan, doneChan, status, runningProcesses)

		// Listen for events fired when a child exits.
		case status := <-doneChan:
			runningProcesses--
			handleDone(procs, runningChan, doneChan, status, runningProcesses)
		}
	}
}
