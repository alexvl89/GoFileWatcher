package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kardianos/service"
)

var logger service.Logger

func main() {
    svcConfig := &service.Config{
        Name:        "GoFileWatcher",
        DisplayName: "Go File Watcher Service",
        Description: "This service watches a directory and copies files with specific extensions to a target directory.",
    }

    prg := &program{}
    s, err := service.New(prg, svcConfig)
    if err != nil {
        log.Fatal(err)
    }

    logger, err = s.Logger(nil)
    if err != nil {
        log.Fatal(err)
    }

    // Process command line arguments
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
    }
}
