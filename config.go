/*
 * Copyright (c) 2016, Randy Westlund. All rights reserved.
 * This code is under the BSD-2-Clause license.
 *
 * This is the main file. Run it to launch the application.
 */
package main

import (
	"github.com/BurntSushi/toml"
	"log"
)

func ParseConfigFile(filename string) *Config {
	var config Config
	_, err := toml.DecodeFile(filename, &config)
	if err != nil {
		log.Fatal("Failed to parse config file", config_file, "\n", err)
	}
	return &config
}
