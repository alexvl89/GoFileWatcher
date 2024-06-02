package main

import (
	"encoding/json"
	"os"
)

type Config struct {
    WatchDirectory  string   `json:"watch_directory"`
    TargetDirectory string   `json:"target_directory"`
    FileExtensions  []string `json:"file_extensions"`
}

func loadConfig(configFile string) (Config, error) {
    var config Config
    file, err := os.Open(configFile)
    if err != nil {
        return config, err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    err = decoder.Decode(&config)
    return config, err
}
