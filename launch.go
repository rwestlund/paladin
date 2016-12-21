/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
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

/* Take a ProcessConfig and Global object. Launch the process, wait for it to
 * exit, and signal completion by returning a Status struct on the completion
 * channel, which includes a possible error.
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

	/* If there's an output file for stderr that isn't the same as stdout. */
	if pc.Stderr != "" && pc.Stderr != pc.Stdout {
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
		/* Follow stdout unless stderr is set. */
		cmd.Stderr = cmd.Stdout
	}

	/* Set user and group. */
	uid, gid, err := GetUidGid(pc.User, pc.Group)
	if err != nil {
		g.DoneChan <- LaunchStatus{Name: pc.Name, Err: err}
		return
	}

	if uid != 0 || gid != 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential =
			&syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}

	/* Fire off the child process, then wait for it to complete. */
	var start_time = time.Now()
	err = cmd.Start()
	if err != nil {
		log.Println("Process", pc.Name, "\tfailed to start", err.Error())
		g.DoneChan <- LaunchStatus{Name: pc.Name, Err: err}
		return
	}

	/* Signal that the process is running. */
	g.RunningChan <- LaunchStatus{Name: pc.Name, Pid: cmd.Process.Pid}

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

func GetUidGid(u string, g string) (uint32, uint32, error) {
	var uid int
	var gid int
	/* Set user and group. */
	if g != "" {
		group_obj, err := user.LookupGroup(g)
		if err != nil {
			log.Println("Failed to lookup gid", err.Error())
			return 0, 0, err
		}
		gid, err = strconv.Atoi(group_obj.Gid)
		if err != nil {
			log.Println("Failed to parse gid", err.Error())
			return 0, 0, err
		}
	}
	if u != "" {
		user_obj, err := user.Lookup(u)
		if err != nil {
			log.Println("Failed to lookup uid", err.Error())
			return 0, 0, err
		}
		uid, err = strconv.Atoi(user_obj.Uid)
		if err != nil {
			log.Println("Failed to parse uid", err.Error())
			return 0, 0, err
		}
	}
	return uint32(uid), uint32(gid), nil
}
