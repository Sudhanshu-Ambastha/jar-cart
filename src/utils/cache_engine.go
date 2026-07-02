package utils

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/charmbracelet/log"
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

var onlineStatus *bool

func IsOnline() bool {
    if onlineStatus != nil {
        return *onlineStatus
    }
    client := http.Client{Timeout: 2 * time.Second}
    _, err := client.Get("https://repo1.maven.org/maven2")
    status := (err == nil)
    onlineStatus = &status
    return status
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

	log.Info("Scanning lib/ for cleanup", "total_files", len(files))

	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || filepath.Ext(fileName) != ".jar" {
			continue
		}

		if !expectedFiles[fileName] {
			log.Warn("Removing unused dependency", "file", fileName)
			fullPath := filepath.Join(libDir, fileName)
			if err := os.Remove(fullPath); err != nil {
				log.Error("Failed to remove file", "file", fileName, "error", err)
			}
		} else {
			log.Debug("Keeping dependency", "file", fileName)
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

func ResolveParallelDependencies(projectDir string, dependencies []models.Dependency, resolveTransitives bool) (map[string]LockEntry, error) {
	homeDir, _ := os.UserHomeDir()
	globalCacheDir := filepath.Join(homeDir, ".jar-cart", "cache")

	fullDepMap := make(map[string]models.Dependency)
	queue := append([]models.Dependency{}, dependencies...)

	for len(queue) > 0 {
		dep := queue[0]
		queue = queue[1:]
		key := fmt.Sprintf("%s:%s", dep.Group, dep.Library)
		
		if _, exists := fullDepMap[key]; !exists {
			fullDepMap[key] = dep
			if resolveTransitives && IsOnline() {
				transitives, err := GetTransitiveDependencies(dep)
				if err == nil {
					queue = append(queue, transitives...)
				} else {
					log.Warn("Could not resolve transitives (POM missing)", "dep", key)
				}
			} else if !resolveTransitives {
				log.Debug("Shallow mode: Skipping transitive resolution", "dep", key)
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
				err := processDependencyExecution(task.Dep, globalCacheDir, filepath.Join(task.ProjectDir, "lib"))
				if err != nil {
					log.Error("Failed to process dependency", "dep", task.Dep.Library, "err", err)
					task.Error = err
				}
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
			continue 
		}
		
		jarName := fmt.Sprintf("%s-%s.jar", res.Dep.Library, res.Dep.Version)
		jarPath := filepath.Join(projectDir, "lib", jarName)
		
		if _, err := os.Stat(jarPath); err == nil {
			hash, _ := CalculateSHA256(jarPath)
			info, _ := os.Stat(jarPath)
			lockEntries[res.Dep.Group+":"+res.Dep.Library] = LockEntry{
				Path:   filepath.Join("lib", jarName),
				Size:   info.Size(),
				SHA256: hash,
			}
		}
	}
	return lockEntries, nil
}

func downloadWithMirrors(dep models.Dependency, targetPath string) error {
	mirrors := []string{
		"https://repo1.maven.org/maven2",
		"https://repo.maven.apache.org/maven2",
		"https://maven.aliyun.com/repository/public",
	}
	
	groupURLPath := strings.ReplaceAll(dep.Group, ".", "/")
	jarName := fmt.Sprintf("%s-%s.jar", dep.Library, dep.Version)
	
	for _, base := range mirrors {
		url := fmt.Sprintf("%s/%s/%s/%s/%s", base, groupURLPath, dep.Library, dep.Version, jarName)
		log.Info("Attempting download", "mirror", base)
		if err := streamAndCacheAsset(url, targetPath); err == nil {
			return nil
		}
	}
	return fmt.Errorf("failed to download from all mirrors")
}

func ResolveCacheToCoordinate(filename string) string {
    base := strings.TrimSuffix(filename, ".jar")
    parts := strings.Split(base, "-")
    version := parts[len(parts)-1]
    artifact := strings.Join(parts[:len(parts)-1], "-")
    return fmt.Sprintf("org.xerial:%s:%s", artifact, version)
}

func linkFromCacheToLib(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create lib dir: %w", err)
	}
	_ = os.Remove(dst)
	err := os.Link(src, dst)
	if err != nil {
		return fallbackFileCopy(src, dst)
	}
	log.Info("Linked dependency to lib/", "file", filepath.Base(dst))
	return nil
}

func processDependencyExecution(dep models.Dependency, globalCacheDir, localLibDir string) error {
    jarName := fmt.Sprintf("%s-%s.jar", dep.Library, dep.Version)
    targetLocalJarPath := filepath.Join(localLibDir, jarName)
    if isFileValid(targetLocalJarPath, "") {
        return nil
    }

    sanitizedGroup := strings.ReplaceAll(dep.Group, ".", string(os.PathSeparator))
    cacheFolder := filepath.Join(globalCacheDir, sanitizedGroup)
    cachedJarPath := filepath.Join(cacheFolder, jarName)
    
    if isFileValid(cachedJarPath, "") {
        return linkFromCacheToLib(cachedJarPath, targetLocalJarPath)
    }
    if !IsOnline() {
        return fmt.Errorf("network offline and '%s' not in cache", jarName)
    }

    os.MkdirAll(cacheFolder, 0755)
    if err := downloadWithMirrors(dep, cachedJarPath); err != nil {
        return err
    }

    return linkFromCacheToLib(cachedJarPath, targetLocalJarPath)
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
	defer out.Close()
	
	pw := &progressWriter{
		total:     resp.ContentLength,
		lastPrint: time.Now(),
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	
	if pw.total > 0 {
		fmt.Println() 
	}

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