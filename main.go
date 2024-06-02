package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/service"
)

var logger service.Logger

func initLogFile(exePath string) {
    logFile := filepath.Join(exePath, "service.log")
    f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Failed to open log file: %v", err)
    }
    log.SetOutput(f)
}

func main() {
    exePath, err := os.Executable()
    if err != nil {
        log.Fatalf("Failed to get executable path: %v", err)
    }
    exeDir := filepath.Dir(exePath)
    initLogFile(exeDir)
    log.Println("Service is starting...")

    configFilePath := filepath.Join(exeDir, "config.json")

    svcConfig := &service.Config{
        Name:        "GoFileWatcher",
        DisplayName: "Go File Watcher Service",
        Description: "This service watches a directory and copies files with specific extensions to a target directory.",
    }

    prg := &program{configFilePath: configFilePath}
    s, err := service.New(prg, svcConfig)
    if err != nil {
        log.Fatalf("Failed to create service: %v", err)
    }

    logger, err = s.Logger(nil)
    if err != nil {
        log.Fatalf("Failed to create service logger: %v", err)
    }

    if len(os.Args) > 1 {
        cmd := os.Args[1]
        switch cmd {
        case "install":
            err = s.Install()
            if err != nil {
                log.Fatalf("Failed to install service: %v", err)
            }
            fmt.Println("Service installed")
        case "uninstall":
            err = s.Uninstall()
            if err != nil {
                log.Fatalf("Failed to uninstall service: %v", err)
            }
            fmt.Println("Service uninstalled")
        case "start":
            err = s.Start()
            if err != nil {
                log.Fatalf("Failed to start service: %v", err)
            }
            fmt.Println("Service started")
        case "stop":
            err = s.Stop()
            if err != nil {
                log.Fatalf("Failed to stop service: %v", err)
            }
            fmt.Println("Service stopped")
        default:
            fmt.Println("Invalid command")
        }
        return
    }

    err = s.Run()
    if err != nil {
        logger.Error(err)
        log.Fatalf("Service failed to run: %v", err)
    }
}
