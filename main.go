/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */
package main

import (
    "log"
    "strings"
    "os"
    "os/exec"
    "github.com/BurntSushi/toml"
)

/* The location of the configuration file to read. */
const config_file = "superv.conf"

/* A process definition, as read directly from the config file. */
type Process struct {
    /* A human-readable tag for process. */
    Name           string
    /* The path to the actual executable to run. */
    Path           string
    /* A string with arguments to be passed to the process. */
    Args           string
    /* Filenames for writing stdout and stderr. Defaults to /dev/null. */
    Stdout         string
    Stderr         string
    /* The number of milliseconds to wait before restarting a process. */
    Restart_delay  uint64
}

/* The config file definition. Currently, there are only [[process]] blocks,
 * but global options may be added in the future.
 */
type Config struct {
    /* This must be named after the [[process]] block in the config file. */
    Process      []Process
}

func main() {
    log.Println("Starting superv")
    /* Read the config file. */
    var config Config
    _, err := toml.DecodeFile(config_file, &config)
    if err != nil {
        log.Fatal("Failed to parse config file ", config_file, "\n", err)
    }

    /* Listen on this channel to know when all goroutines are done. */
    var done = make(chan bool)
    for _, p := range config.Process {
        log.Println("Launching ", p.Name)
        /* Launch the process in a new goroutine. */
        go launch(p, done)
    }
    /* Listen for one done signal from each goroutine. */
    for range config.Process {
        <-done
        log.Println("A process finished");
    }
    log.Println("Exiting");
}

/* Take a process and notification channel. Launch the process, wait for it to
 * exit, and signal completion.
 */
func launch(p Process, done chan bool) {
    /* Convert p.args to a slice, so the process gets separate arguments. */
    var cmd = exec.Command(p.Path, squeeze(strings.Split(p.Args, " "))...)
    /* Use the parent's stdout and stderr. */
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()
    cmd.Wait()
    /* Signal completion. */
    done <- true
}

/* Remove empty strings from a slice of strings. Returns a new slice. */
func squeeze(string_slice []string) []string {
    var ss []string
    for _, s := range string_slice {
        if s != "" {
            ss = append(ss, s)
        }
    }
    return ss
}
