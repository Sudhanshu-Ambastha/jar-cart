package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func ListJDKs() {
	homeDir, _ := os.UserHomeDir()
	jdkRoot := filepath.Join(homeDir, ".jar-cart", "jdks")
	
	fmt.Println("☕ Installed Java Runtimes:")
	entries, err := os.ReadDir(jdkRoot)
	if err != nil {
		fmt.Println(" No JDKs found.")
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf(" - Java %s\n", entry.Name())
		}
	}
}

func RemoveJDK(version string) error {
	homeDir, _ := os.UserHomeDir()
	targetPath := filepath.Join(homeDir, ".jar-cart", "jdks", version)
	
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("JDK %s not found", version)
	}
	
	log.Info("Removing JDK...", "version", version)
	return os.RemoveAll(targetPath)
}

func ListCache() {
	homeDir, _ := os.UserHomeDir()
	cacheRoot := filepath.Join(homeDir, ".jar-cart", "cache")
	
	fmt.Println("📦 Global Cache Inventory:")
	filepath.Walk(cacheRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if !info.IsDir() && filepath.Ext(path) == ".jar" {
			rel, _ := filepath.Rel(cacheRoot, path)
			fmt.Printf(" - %s (%.2f MB)\n", rel, float64(info.Size())/(1024*1024))
		}
		return nil
	})
}

func RemoveCachedItem(target string) error {
	homeDir, _ := os.UserHomeDir()
	cacheRoot := filepath.Join(homeDir, ".jar-cart", "cache")	
	var targetsToRemove []string

	err := filepath.Walk(cacheRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil { return nil }
		if strings.Contains(path, target) {
			targetsToRemove = append(targetsToRemove, path)
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("error walking cache: %w", err)
	}

	if len(targetsToRemove) == 0 {
		return fmt.Errorf("no items matching '%s' found in cache", target)
	}

	for _, path := range targetsToRemove {
		log.Info("Removing item...", "path", path)
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	
	return nil
}