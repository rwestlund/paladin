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
        log.Fatal("Failed to parse config file", config_file, "\n", err)
    }

    /* Listen on this channel to know when all goroutines are done. */
    var done = make(chan int)

    /* Launch all processes. */
    for i, p := range config.Process {
        log.Println("Process", p.Name, "launching")
        /* Launch the process in a new goroutine. */
        go launch(p, i, done)
    }

    /* Listen for one done signal from each goroutine. */
    for range config.Process {
        var index = <-done
        log.Println("Process", config.Process[index].Name, "finished")
    }
    time.Sleep(10)
    log.Println("Exiting")
}

/* Take a process, an id, and notification channel. Launch the process, wait
 * for it to exit, and signal completion by returning the id on the channel.
 */
func launch(p Process, id int, done chan int) {
    /* Convert p.args to a slice, so the process gets separate arguments. */
    var cmd = exec.Command(p.Path, squeeze(strings.Split(p.Args, " "))...)

    /* If there's an output file for stdout specified, use it. */
    if p.Stdout != "" {
        stdout_file, err := os.OpenFile(p.Stdout,
                    os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0664)
        if err != nil {
            log.Println("Failed to open log file", p.Stdout, "\n", err)
        }
        defer stdout_file.Close()
        cmd.Stdout = stdout_file
    /* If not, use /dev/null. */
    } else {
        cmd.Stdout = nil
    }

    /* If there's an output file for stderr specified, use it. */
    if p.Stderr != "" {
        stderr_file, err := os.OpenFile(p.Stderr,
                    os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0664)
        if err != nil {
            log.Println("Failed to open log file", p.Stderr, "\n", err)
        }
        defer stderr_file.Close()
        cmd.Stderr = stderr_file
    /* If not, use /dev/null. */
    } else {
        cmd.Stderr = nil
    }

    /* Fire off the chiled process, then wait for it to complete. */
    cmd.Start()
    cmd.Wait()
    /* Signal completion. */
    done <- id
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
