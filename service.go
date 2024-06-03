package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kardianos/service"
)

type program struct {
    configFilePath string
    config         Config
    watcher        *fsnotify.Watcher
    mu             sync.Mutex
}

func (p *program) Start(s service.Service) error {
    // Load config
    var err error
    p.config, err = loadConfig(p.configFilePath)
    if err != nil {
        log.Printf("Error loading config: %v", err)
        return err
    }

    // Mount network directory
    if err = mountNetworkDirectory(p.config.TargetDirectory, p.config.NetworkUser, p.config.NetworkPassword); err != nil {
        log.Printf("Error mounting network directory: %v", err)
        return err
    }

    // Validate directories
    if _, err := os.Stat(p.config.WatchDirectory); os.IsNotExist(err) {
        log.Printf("Watch directory does not exist: %s", p.config.WatchDirectory)
        return fmt.Errorf("watch directory does not exist: %s", p.config.WatchDirectory)
    }
    if _, err := os.Stat(p.config.TargetDirectory); os.IsNotExist(err) {
        log.Printf("Target directory does not exist: %s", p.config.TargetDirectory)
        return fmt.Errorf("target directory does not exist: %s", p.config.TargetDirectory)
    }

    // Create watcher
    p.watcher, err = fsnotify.NewWatcher()
    if err != nil {
        log.Printf("Error creating watcher: %v", err)
        return err
    }
    go p.run()
    log.Println("Service started successfully")
    return p.watcher.Add(p.config.WatchDirectory)
}

func (p *program) run() {
    for {
        select {
        case event, ok := <-p.watcher.Events:
            if !ok {
                return
            }
            if event.Op&fsnotify.Create == fsnotify.Create {
                log.Printf("Detected new file: %s", event.Name)
                go p.handleCreateEvent(event)
            }
        case err, ok := <-p.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Watcher error: %v", err)
        }
    }
}

func (p *program) handleCreateEvent(event fsnotify.Event) {
    for _, ext := range p.config.FileExtensions {
        if strings.HasSuffix(event.Name, ext) {
            go func(fileName string) {
                p.mu.Lock()
                defer p.mu.Unlock()

                if waitForFile(fileName, 10*time.Second) {
                    targetPath := filepath.Join(p.config.TargetDirectory, filepath.Base(fileName))
                    log.Printf("Copying file from %s to %s", fileName, targetPath)
                    if err := copyFileWithRetries(fileName, targetPath, 5); err != nil {
                        log.Printf("Failed to copy file: %s", err)
                    } else {
                        log.Printf("Copied file: %s to %s", fileName, targetPath)
                    }
                } else {
                    log.Printf("File %s did not stabilize in time", fileName)
                }
            }(event.Name)
            break
        }
    }
}

func waitForFile(fileName string, timeout time.Duration) bool {
    ticker := time.NewTicker(500 * time.Millisecond)
    defer ticker.Stop()

    timeoutTimer := time.NewTimer(timeout)
    defer timeoutTimer.Stop()

    var lastSize int64 = -1
    for {
        select {
        case <-ticker.C:
            fileInfo, err := os.Stat(fileName)
            if err != nil {
                log.Printf("Error stating file: %s", err)
                return false
            }

            currentSize := fileInfo.Size()
            if currentSize == lastSize {
                return true
            }
            lastSize = currentSize

        case <-timeoutTimer.C:
            return false
        }
    }
}

func mountNetworkDirectory(targetDir, username, password string) error {
    // Unmount all previous connections to the network server
    unmountAllCmd := exec.Command("net", "use", "*", "/delete", "/yes")
    unmountAllOutput, unmountAllErr := unmountAllCmd.CombinedOutput()
    if unmountAllErr != nil {
        log.Printf("Warning: Unmounting all network connections failed: %v. Output: %s", unmountAllErr, unmountAllOutput)
    }

    // Unmount the specific network directory if it is already mounted
    unmountCmd := exec.Command("net", "use", targetDir, "/delete")
    unmountOutput, unmountErr := unmountCmd.CombinedOutput()
    if unmountErr != nil {
        log.Printf("Warning: Unmounting network directory failed: %v. Output: %s", unmountErr, unmountOutput)
    }

    // Mount the network directory
    cmd := exec.Command("net", "use", targetDir, "/user:"+username, password)
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Error mounting network directory. Command: %s, Output: %s", cmd.String(), output)
        return fmt.Errorf("failed to mount network directory: %w, output: %s", err, output)
    }
    return nil
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
    log.Println("Service stopping")
    return p.watcher.Close()
}
