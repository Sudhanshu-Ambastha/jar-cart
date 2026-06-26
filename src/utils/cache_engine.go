package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type SyncTask struct {
	Dep        models.Dependency
	ProjectDir string
	Error      error
}

func isFileValid(path string, expectedHash string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if expectedHash == "" {
		return true
	}
	actualHash, err := CalculateSHA256(path)
	return err == nil && actualHash == expectedHash
}

func CleanupLibDir(projectDir string, expectedEntries map[string]LockEntry) error {
	libDir := filepath.Join(projectDir, "lib")
	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		return nil 
	}

	files, err := os.ReadDir(libDir)
	if err != nil {
		return err
	}

	expectedFiles := make(map[string]bool)
	for _, entry := range expectedEntries {
		expectedFiles[filepath.Base(entry.Path)] = true
	}

	fmt.Printf("🧹 Scanning lib/ for cleanup (Total files: %d)\n", len(files))

	for _, file := range files {
		fileName := file.Name()
		
		if file.IsDir() || filepath.Ext(fileName) != ".jar" {
			continue
		}

		if !expectedFiles[fileName] {
			fmt.Printf("🗑️ Removing unused: %s\n", fileName)
			fullPath := filepath.Join(libDir, fileName)
			if err := os.Remove(fullPath); err != nil {
				fmt.Printf("❌ Failed to remove %s: %v\n", fileName, err)
			}
		} else {
			fmt.Printf("✅ Keeping: %s\n", fileName)
		}
	}
	return nil
}

func GetTransitiveDependencies(dep models.Dependency) ([]models.Dependency, error) {
	groupPath := strings.ReplaceAll(dep.Group, ".", "/")
	url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.pom",
		groupPath, dep.Library, dep.Version, dep.Library, dep.Version)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not fetch POM for %s:%s", dep.Group, dep.Library)
	}
	defer resp.Body.Close()

	var pom models.Pom
	if err := xml.NewDecoder(resp.Body).Decode(&pom); err != nil {
		return nil, err
	}

	var transitives []models.Dependency
	for _, d := range pom.Dependencies {
		if d.Scope == "test" || d.Scope == "provided" {
			continue
		}
		if d.Version != "" {
			transitives = append(transitives, models.Dependency{
				Group: d.GroupID, Library: d.ArtifactID, Version: d.Version,
			})
		}
	}
	return transitives, nil
}

func ResolveParallelDependencies(projectDir string, dependencies []models.Dependency) (map[string]LockEntry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}
	globalCacheDir := filepath.Join(homeDir, ".jar-cart", "cache")

	fullDepMap := make(map[string]models.Dependency)
	queue := dependencies

	for len(queue) > 0 {
		dep := queue[0]
		queue = queue[1:]
		key := fmt.Sprintf("%s:%s", dep.Group, dep.Library)
		if _, exists := fullDepMap[key]; !exists {
			fullDepMap[key] = dep
			if transitives, err := GetTransitiveDependencies(dep); err == nil {
				queue = append(queue, transitives...)
			}
		}
	}

	numWorkers := runtime.NumCPU()
	tasksChan := make(chan *SyncTask, len(fullDepMap))
	resultsChan := make(chan *SyncTask, len(fullDepMap))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksChan {
				task.Error = processDependencyExecution(task.Dep, globalCacheDir, filepath.Join(task.ProjectDir, "lib"))
				resultsChan <- task
			}
		}()
	}

	for _, dep := range fullDepMap {
		tasksChan <- &SyncTask{Dep: dep, ProjectDir: projectDir}
	}
	close(tasksChan)
	wg.Wait()
	close(resultsChan)

	lockEntries := make(map[string]LockEntry)
	for res := range resultsChan {
		if res.Error != nil {
			return nil, res.Error
		}
		jarName := fmt.Sprintf("%s-%s.jar", res.Dep.Library, res.Dep.Version)
		jarPath := filepath.Join(projectDir, "lib", jarName)
		hash, _ := CalculateSHA256(jarPath)
		info, _ := os.Stat(jarPath)
		lockEntries[res.Dep.Group+":"+res.Dep.Library] = LockEntry{
			Path:   filepath.Join("lib", jarName),
			Size:   info.Size(),
			SHA256: hash,
		}
	}
	return lockEntries, nil
}

func processDependencyExecution(dep models.Dependency, globalCacheDir, localLibDir string) error {
	jarName := fmt.Sprintf("%s-%s.jar", dep.Library, dep.Version)
	coordHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", dep.Group, dep.Library, dep.Version)))
	hashStr := hex.EncodeToString(coordHash[:12])
	cachePackageDir := filepath.Join(globalCacheDir, hashStr)
	cachedJarPath := filepath.Join(cachePackageDir, jarName)
	targetLocalJarPath := filepath.Join(localLibDir, jarName)

	if isFileValid(targetLocalJarPath, "") {
		return nil
	}

	if !isFileValid(cachedJarPath, "") {
		fmt.Printf("📥 Downloading: %s\n", jarName)
		_ = os.MkdirAll(cachePackageDir, 0755)
		groupURLPath := strings.ReplaceAll(dep.Group, ".", "/")
		url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.jar",
			groupURLPath, dep.Library, dep.Version, dep.Library, dep.Version)

		if err := streamAndCacheAsset(url, cachedJarPath); err != nil {
			return err
		}
	}

	fmt.Printf("📦 Cached: %s\n", jarName)

	_ = os.Remove(targetLocalJarPath)
	err := os.Link(cachedJarPath, targetLocalJarPath)
	if err != nil {
		return fallbackFileCopy(cachedJarPath, targetLocalJarPath)
	}
	return nil
}

func streamAndCacheAsset(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("maven error: %s", resp.Status)
	}

	tmpPath := targetPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, targetPath)
}

func fallbackFileCopy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}