package utils

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

func WatchAndRun(input string, appArgs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to initialize watcher", "error", err)
	}
	defer watcher.Close()

	fileHashes := make(map[string][32]byte)
	log.Info("Watching for changes (Content-Verified)...", "args", appArgs)
	RunProject(input, appArgs)

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (path == ".jar-cart" || path == ".git") {
			return filepath.SkipDir
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".java") || path == "jar-cart.json") {
			_ = watcher.Add(path)
			hash, _ := GetFileHash(path) 
			fileHashes[path] = hash
		}
		return nil
	})

	var lastRun time.Time

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op != fsnotify.Write {
				continue
			}

			if time.Since(lastRun) < 500*time.Millisecond {
				continue
			}

			newHash, err := GetFileHash(event.Name) 
			if err == nil && newHash == fileHashes[event.Name] {
				continue
			}

			fileHashes[event.Name] = newHash
			lastRun = time.Now()

			log.Info("Actual change detected, recompiling...", "file", event.Name)
			go RunProject(input, appArgs)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("Watcher error", "error", err)
		}
	}
}