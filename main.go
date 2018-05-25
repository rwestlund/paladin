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
	var g global = global{}
	// Initialize pointer types.
	g.RunningChan = make(chan launchStatus)
	g.DoneChan = make(chan launchStatus)
	g.Procs = make(map[string]*process)
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

	// Build the global object.
	for _, pc := range config.Process {
		g.Procs[pc.Name] = &process{Config: pc}
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
				for i := range g.Procs {
					if g.Procs[i].Running {
						log.Println("killing", g.Procs[i].Config.Name)
						unix.Kill(g.Procs[i].Status.Pid, unix.SIGTERM)
					}
				}
				log.Println("done")
				os.Exit(0)
			case unix.SIGINFO:
				fmt.Println("The following children are running:")
				fmt.Fprintln(writer, "Process\tRunning\tPath\tArgs\tUser\tGroup")
				fmt.Fprintln(writer, "-------\t-------\t----\t----\t----\t-----")
				for i := range g.Procs {
					fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n", g.Procs[i].Config.Name,
						printYN(g.Procs[i].Running), g.Procs[i].Config.Path,
						g.Procs[i].Config.Args, g.Procs[i].Config.User, g.Procs[i].Config.Group)
					writer.Flush()
				}
			}
		}
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
			g.Procs[status.Name].Status = status
			handleRunning(&g, status)

		// Listen for events fired when a child exits.
		case status := <-g.DoneChan:
			g.RunningProcesses--
			handleDone(&g, status)
		}
	}
}
