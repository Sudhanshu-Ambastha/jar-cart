package utils

import (
	"context"
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
	"golang.org/x/sync/errgroup"
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

	if info.Size() < 100 {
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

	expectedFileNames := make(map[string]bool)
	for _, entry := range expectedEntries {
		expectedFileNames[filepath.Base(entry.Path)] = true
	}

	for _, file := range files {
		fileName := file.Name()
		if file.IsDir() || filepath.Ext(fileName) != ".jar" {
			continue
		}
		if !expectedFileNames[fileName] {
			log.Warn("Removing stale or extraneous dependency", "file", fileName)
			fullPath := filepath.Join(libDir, fileName)
			if err := os.Remove(fullPath); err != nil {
				log.Error("Failed to remove file", "file", fileName, "error", err)
				return err
			}
		}
	}
	
	log.Info("Cleanup complete. lib/ is in sync with manifest.")
	return nil
}

func ResolveParallelDependencies(targetLibDir string, dependencies []models.Dependency, shouldResolve bool) (map[string]LockEntry, error) {
	homeDir, _ := os.UserHomeDir()
	globalCacheDir := filepath.Join(homeDir, ".jar-cart", "cache")
	var resolvedDeps []models.Dependency
	var err error

	if shouldResolve {
		resolvedDeps, err = ResolveUsingGoogle(dependencies)
	} else {
		resolvedDeps = dependencies
	}

	if err != nil {
		return nil, err
	}

	g, _ := errgroup.WithContext(context.Background())
	g.SetLimit(runtime.NumCPU()) 

	var mu sync.Mutex
	lockEntries := make(map[string]LockEntry)

	for _, dep := range resolvedDeps {
		dep := dep 
		g.Go(func() error {
			err := processDependencyExecution(dep, globalCacheDir, targetLibDir)
			if err != nil {
				return err 
			}

			jarName := fmt.Sprintf("%s-%s.jar", dep.Library, dep.Version)
			jarPath := filepath.Join(targetLibDir, jarName)
			var fileSize int64
			if info, err := os.Stat(jarPath); err == nil {
				fileSize = info.Size()
			}
			
			hash, _ := CalculateSHA256(jarPath)

			mu.Lock()
			lockEntries[dep.Group+":"+dep.Library] = LockEntry{
				Path:   jarPath,
				Size:   fileSize,
				SHA256: hash,
			}
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	
	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}
	
	return out.Close()
}

func linkFromCacheToLib(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create lib dir: %w", err)
	}
	_ = os.Remove(dst)
	if err := copyFile(src, dst); err != nil {
		return err
	}
	
	log.Info("Linked dependency to lib/", "file", filepath.Base(dst))
	return nil
}

func processDependencyExecution(dep models.Dependency, globalCacheDir, localLibDir string) error {
	jarName := fmt.Sprintf("%s-%s.jar", dep.Library, dep.Version)
	targetDir := localLibDir
	if filepath.Base(targetDir) != "lib" {
		targetDir = filepath.Join(localLibDir, "lib")
	}

	targetLocalJarPath := filepath.Join(targetDir, jarName)

	if isFileValid(targetLocalJarPath, "") {
		return nil
	}

	sanitizedGroup := strings.ReplaceAll(dep.Group, ".", string(os.PathSeparator))
	cacheFolder := filepath.Join(globalCacheDir, sanitizedGroup)
	cachedJarPath := filepath.Join(cacheFolder, jarName)

	if !isFileValid(cachedJarPath, "") {
		if !IsOnline() {
			return fmt.Errorf("network offline and artifact not in cache: %s", jarName)
		}
		os.MkdirAll(cacheFolder, 0755)
		if err := downloadWithMirrors(dep, cachedJarPath); err != nil {
			return err
		}
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