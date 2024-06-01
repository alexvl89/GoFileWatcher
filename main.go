package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
)

type Config struct {
    WatchDirectory string   `json:"watch_directory"`
    TargetDirectory string  `json:"target_directory"`
    FileExtensions  []string `json:"file_extensions"`
}

type program struct {
    config Config
    watcher *fsnotify.Watcher
}

var logger service.Logger

func (p *program) Start(s service.Service) error {
    // Load config
    configFile, err := os.Open("config.json")
    if err != nil {
        return err
    }
    defer configFile.Close()
    jsonParser := json.NewDecoder(configFile)
    if err = jsonParser.Decode(&p.config); err != nil {
        return err
    }

    // Validate directories
    if _, err := os.Stat(p.config.WatchDirectory); os.IsNotExist(err) {
        return fmt.Errorf("Watch directory does not exist: %s", p.config.WatchDirectory)
    }
    if _, err := os.Stat(p.config.TargetDirectory); os.IsNotExist(err) {
        return fmt.Errorf("Target directory does not exist: %s", p.config.TargetDirectory)
    }

    // Create watcher
    p.watcher, err = fsnotify.NewWatcher()
    if err != nil {
        return err
    }
    go p.run()
    return p.watcher.Add(p.config.WatchDirectory)
}

func (p *program) run() {
    logFile, err := os.OpenFile("service.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatal(err)
    }
    defer logFile.Close()
    log.SetOutput(logFile)

    for {
        select {
        case event, ok := <-p.watcher.Events:
            if !ok {
                return
            }
            if event.Op&fsnotify.Create == fsnotify.Create {
                p.handleCreateEvent(event)
            }
        case err, ok := <-p.watcher.Errors:
            if !ok {
                return
            }
            log.Println("error:", err)
        }
    }
}

func (p *program) handleCreateEvent(event fsnotify.Event) {
    for _, ext := range p.config.FileExtensions {
        if strings.HasSuffix(event.Name, ext) {
            targetPath := filepath.Join(p.config.TargetDirectory, filepath.Base(event.Name))
            if err := copyFile(event.Name, targetPath); err != nil {
                log.Printf("Failed to copy file: %s", err)
            } else {
                log.Printf("Copied file: %s to %s", event.Name, targetPath)
            }
            break
        }
    }
}

func copyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destinationFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destinationFile.Close()

    if _, err := io.Copy(destinationFile, sourceFile); err != nil {
        return err
    }
    return nil
}

func (p *program) Stop(s service.Service) error {
    return p.watcher.Close()
}

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

    err = s.Run()
    if err != nil {
        logger.Error(err)
    }
}
