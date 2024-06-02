package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
)

type program struct {
    config  Config
    watcher *fsnotify.Watcher
    mu      sync.Mutex
}

func (p *program) Start(s service.Service) error {
    // Load config
    var err error
    p.config, err = loadConfig("config.json")
    if err != nil {
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
                go p.handleCreateEvent(event)
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
            p.mu.Lock()
            defer p.mu.Unlock()

            targetPath := filepath.Join(p.config.TargetDirectory, filepath.Base(event.Name))
            if err := copyFileWithRetries(event.Name, targetPath, 5); err != nil {
                log.Printf("Failed to copy file: %s", err)
            } else {
                log.Printf("Copied file: %s to %s", event.Name, targetPath)
            }
            break
        }
    }
}

func copyFileWithRetries(src, dst string, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = copyFile(src, dst)
        if err == nil {
            return nil
        }
        log.Printf("Retrying to copy file: %s, attempt %d", src, i+1)
        time.Sleep(2 * time.Second)
    }
    return fmt.Errorf("failed to copy file %s after %d retries: %w", src, maxRetries, err)
}

func (p *program) Stop(s service.Service) error {
    return p.watcher.Close()
}
