package main

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

func newWatcher() (*fsnotify.Watcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
        return nil, err
    }
    return watcher, nil
}
