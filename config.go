package main

import (
    "log"
    "github.com/BurntSushi/toml"
)

func ParseConfigFile(filename string) *Config {
    var config Config
    _, err := toml.DecodeFile(filename, &config)
    if err != nil {
        log.Fatal("Failed to parse config file", config_file, "\n", err)
    }
    return &config
}