package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func WatchAndRun(input string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("👀 Watching for changes in 'src/' and 'jar-cart.json'... (Press Ctrl+C to stop)")
	RunProject(input)
	filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})
	watcher.Add("jar-cart.json")

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok { return }
			isJava := strings.HasSuffix(event.Name, ".java")
			isManifest := strings.Contains(event.Name, "jar-cart.json")
			
			if (isJava || isManifest) && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
				if isManifest {
					fmt.Printf("\n⚙️ Manifest change detected! Syncing and reloading with new configuration...\n")
				} else {
					fmt.Printf("\n🔄 Change detected in %s. Recompiling...\n", event.Name)
				}
				RunProject(input)
			}
		case err, ok := <-watcher.Errors:
			if !ok { return }
			fmt.Println("❌ Watcher error:", err)
		}
	}
}