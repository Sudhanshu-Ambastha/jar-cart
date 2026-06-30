package utils

import (
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

func getFileHash(path string) ([32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, err
	}
	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res, nil
}

func WatchAndRun(input string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to initialize watcher", "error", err)
	}
	defer watcher.Close()
	fileHashes := make(map[string][32]byte)

	log.Info("Watching for changes (Content-Verified)...")
	RunProject(input)

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && (path == "bin" || path == "dist" || path == ".git") {
			return filepath.SkipDir
		}
		if err == nil && !info.IsDir() && (strings.HasSuffix(path, ".java") || path == "jar-cart.json") {
			watcher.Add(path)
			hash, _ := getFileHash(path)
			fileHashes[path] = hash
		}
		return nil
	})

	var lastRun time.Time
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok { return }
			if event.Op != fsnotify.Write { continue }
			newHash, err := getFileHash(event.Name)
			if err == nil && newHash == fileHashes[event.Name] {
				continue 
			}
			fileHashes[event.Name] = newHash
			if time.Since(lastRun) < 1*time.Second { continue }
			lastRun = time.Now()

			log.Info("Actual change detected, recompiling...", "file", event.Name)
			go RunProject(input)

		case err, ok := <-watcher.Errors:
			if !ok { return }
			log.Error("Watcher error", "error", err)
		}
	}
}