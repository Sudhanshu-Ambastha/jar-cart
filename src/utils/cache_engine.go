package utils

import (
	"crypto/sha256"
	"encoding/hex"
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

func ResolveParallelDependencies(projectDir string, dependencies []models.Dependency) (map[string]LockEntry, error) {
	if len(dependencies) == 0 {
		return nil, nil
	}

	homeDir, _ := os.UserHomeDir()
	globalCacheDir := filepath.Join(homeDir, ".jar-cart", "cache")
	localLibDir := filepath.Join(projectDir, "lib")

	_ = os.MkdirAll(globalCacheDir, 0755)
	_ = os.MkdirAll(localLibDir, 0755)

	numWorkers := runtime.NumCPU()
	tasksChan := make(chan *SyncTask, len(dependencies))
	resultsChan := make(chan *SyncTask, len(dependencies))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksChan {
				task.Error = processDependencyExecution(task.Dep, globalCacheDir, localLibDir)
				resultsChan <- task
			}
		}()
	}

	for _, dep := range dependencies {
		tasksChan <- &SyncTask{Dep: dep, ProjectDir: projectDir}
	}
	close(tasksChan)
	wg.Wait()
	close(resultsChan)

	lockEntries := make(map[string]LockEntry)
	for res := range resultsChan {
		if res.Error != nil {
			return nil, fmt.Errorf("pipeline fault: %v", res.Error)
		}
		jarPath := filepath.Join(localLibDir, fmt.Sprintf("%s-%s.jar", res.Dep.Library, res.Dep.Version))
		hash, _ := CalculateSHA256(jarPath)
		info, _ := os.Stat(jarPath)
		lockEntries[res.Dep.Group+":"+res.Dep.Library] = LockEntry{
			Path:   filepath.Join("lib", filepath.Base(jarPath)),
			Size:   info.Size(),
			SHA256: hash,
		}
	}
	return lockEntries, nil
}

func DownloadJars(dependencies []models.Dependency, libDir string) error {
	if len(dependencies) == 0 {
		return nil
	}
	_, err := ResolveParallelDependencies(".", dependencies)
	return err
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