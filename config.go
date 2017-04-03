/*
 * Copyright (c) 2016-2017, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */

package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

func parseConfigFile(filename string) *configOptions {
	var config configOptions
	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		log.Fatal("Failed to parse config file", configFile, "\n", err)
	}
	return &config
}
