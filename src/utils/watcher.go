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

	fmt.Println("👀 Watching 'src/' for changes... (Press Ctrl+C to stop)")
	RunProject(input)

	filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return watcher.Add(path)
		}
		return nil
	})

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok { return }
			if strings.HasSuffix(event.Name, ".java") {
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Printf("\n🔄 Change detected in %s. Recompiling...\n", event.Name)
					RunProject(input)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok { return }
			fmt.Println("❌ Watcher error:", err)
		}
	}
}