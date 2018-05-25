/*
 * Copyright (c) 2016-2017, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This file contains the logic for spawning child processes.
 */

package main

import (
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Take a processConfig and global object. Launch the process, wait for it to
// exit, and signal completion by returning a status struct on the completion
// channel, which includes a possible error.
func launchProcess(pc processConfig, g *global) {
	log.Println("Process", pc.Name, "\tlaunching")
	// Convert p.args to a slice, so the process gets separate arguments.
	var cmd = exec.Command(pc.Path, squeeze(strings.Split(pc.Args, " "))...)
	// Set the process PGID to not match paladin's, so that it receives
	// signals separately.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Set the current working directory for the process.
	if pc.Cwd != "" {
		cmd.Dir = pc.Cwd
	} else {
		cmd.Dir = "/var/empty"
	}

	// If there's an output file for stdout specified, use it.
	if pc.Stdout != "" {
		var stdoutFile, err = os.OpenFile(pc.Stdout,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			log.Println("Failed to open log file", pc.Stdout, "\n", err)
			g.DoneChan <- launchStatus{Name: pc.Name, Err: err}
			return
		}
		defer stdoutFile.Close()
		cmd.Stdout = stdoutFile
	} else {
		// If not, use /dev/null.
		cmd.Stdout = nil
	}

	// If there's an output file for stderr that isn't the same as stdout.
	if pc.Stderr != "" && pc.Stderr != pc.Stdout {
		var stderrFile, err = os.OpenFile(pc.Stderr,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
		if err != nil {
			log.Println("Failed to open log file", pc.Stderr, "\n", err)
			g.DoneChan <- launchStatus{Name: pc.Name, Err: err}
			return
		}
		defer stderrFile.Close()
		cmd.Stderr = stderrFile
	} else {
		// Follow stdout unless stderr is set.
		cmd.Stderr = cmd.Stdout
	}

	// Set user and group.
	var uid, gid, err = getUIDAndGID(pc.User, pc.Group)
	if err != nil {
		g.DoneChan <- launchStatus{Name: pc.Name, Err: err}
		return
	}

	if uid != 0 || gid != 0 {
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}
	}

	// Fire off the child process, then wait for it to complete.
	var startTime = time.Now()
	err = cmd.Start()
	if err != nil {
		log.Println("Process", pc.Name, "\tfailed to start", err.Error())
		g.DoneChan <- launchStatus{Name: pc.Name, Err: err}
		return
	}

	// Signal that the process is running.
	g.RunningChan <- launchStatus{Name: pc.Name, Pid: cmd.Process.Pid}

	// Wait for it to finish.
	err = cmd.Wait()
	var duration = time.Since(startTime)
	if err != nil {
		log.Println("Process", pc.Name, "\tfailed to run")
		g.DoneChan <- launchStatus{Name: pc.Name, Err: err, Duration: duration}
		return
	}
	// Signal completion.
	g.DoneChan <- launchStatus{Name: pc.Name, Err: nil, Duration: duration}
}

// Get the numeric uid and gid for the given user and group.
func getUIDAndGID(u string, g string) (uint32, uint32, error) {
	var uid int
	var gid int
	// Set user and group.
	if g != "" {
		groupObj, err := user.LookupGroup(g)
		if err != nil {
			log.Println("Failed to lookup gid", err.Error())
			return 0, 0, err
		}
		gid, err = strconv.Atoi(groupObj.Gid)
		if err != nil {
			log.Println("Failed to parse gid", err.Error())
			return 0, 0, err
		}
	}
	if u != "" {
		userObj, err := user.Lookup(u)
		if err != nil {
			log.Println("Failed to lookup uid", err.Error())
			return 0, 0, err
		}
		uid, err = strconv.Atoi(userObj.Uid)
		if err != nil {
			log.Println("Failed to parse uid", err.Error())
			return 0, 0, err
		}
	}
	return uint32(uid), uint32(gid), nil
}
