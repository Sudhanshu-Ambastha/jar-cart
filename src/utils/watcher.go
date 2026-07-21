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

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && (path == ".jar-cart" || path == ".git") {
			return filepath.SkipDir
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".java") || path == "jar-cart.json" || path == "jar-cart.xml") {
			_ = watcher.Add(path)
			hash, _ := GetFileHash(path) 
			fileHashes[path] = hash
		}
		return nil
	})
	if err != nil {
		log.Warn("Encountered error while walking path for watcher", "error", err)
	}

	var lastRun time.Time

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
				continue
			}

			if time.Since(lastRun) < 500*time.Millisecond {
				continue
			}

			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && !info.IsDir() {
					if strings.HasSuffix(event.Name, ".java") || event.Name == "jar-cart.json" || event.Name == "jar-cart.xml" {
						_ = watcher.Add(event.Name)
					}
				}
			}

			newHash, err := GetFileHash(event.Name) 
			if err == nil && newHash == fileHashes[event.Name] {
				continue
			}

			fileHashes[event.Name] = newHash
			lastRun = time.Now()

			log.Info("Actual change detected, re-running project pipeline...", "file", event.Name)
			go RunProject(input, appArgs)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("Watcher error", "error", err)
		}
	}
}

func WatchWorkspace() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to initialize workspace watcher", "error", err)
	}
	defer watcher.Close()

	workspace := LoadWorkspaceManifest()
	excluded := GetExcludedModules()
	excludedMap := make(map[string]bool)
	for _, ex := range excluded {
		excludedMap[ex] = true
	}

	fileHashes := make(map[string][32]byte)
	log.Info("Watching workspace for changes (Multi-Module Content-Verified)...")
	OrchestrateBuild("workspace-init")

	for modName, modConfig := range workspace.Modules {
		if excludedMap[modName] || excludedMap[modConfig.Path] {
			continue
		}

		err := filepath.Walk(modConfig.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() && (path == ".jar-cart" || path == ".git" || path == "lib" || path == "bin") {
				return filepath.SkipDir
			}

			if !info.IsDir() && (strings.HasSuffix(path, ".java") || strings.HasSuffix(path, "jar-cart.json") || strings.HasSuffix(path, "jar-cart.xml")) {
				_ = watcher.Add(path)
				hash, _ := GetFileHash(path)
				fileHashes[path] = hash
			}
			return nil
		})
		if err != nil {
			log.Warn("Failed to walk module path for watching", "module", modName, "error", err)
		}
	}

	var lastRun time.Time

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 {
				continue
			}

			if time.Since(lastRun) < 500*time.Millisecond {
				continue
			}

			if event.Op&fsnotify.Create != 0 {
				if info, statErr := os.Stat(event.Name); statErr == nil && !info.IsDir() {
					if strings.HasSuffix(event.Name, ".java") || strings.HasSuffix(event.Name, "jar-cart.json") || strings.HasSuffix(event.Name, "jar-cart.xml") {
						_ = watcher.Add(event.Name)
					}
				}
			}

			newHash, err := GetFileHash(event.Name)
			if err == nil && newHash == fileHashes[event.Name] {
				continue
			}

			fileHashes[event.Name] = newHash
			lastRun = time.Now()

			log.Info("Workspace file change detected, evaluating dirty modules...", "file", event.Name)
			go OrchestrateBuild(event.Name)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Error("Workspace watcher error", "error", err)
		}
	}
}