package utils

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func AutoCheckUpdate(currentVersion string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}
	timestampFile := filepath.Join(homeDir, ".jar-cart", ".last_update_check")

	if info, err := os.Stat(timestampFile); err == nil {
		if time.Since(info.ModTime()) < 24*time.Hour {
			return
		}
	}

	_ = os.MkdirAll(filepath.Dir(timestampFile), 0755)
	_ = os.WriteFile(timestampFile, []byte(time.Now().String()), 0644)

	go func() {
		resp, err := http.Get("https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest")
		if err != nil {
			return
		}
		defer resp.Body.Close()

		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err == nil {
			if release.TagName != currentVersion {
				fmt.Printf("\n✨ [jar-cart] A new version is available: %s (Current: %s)\n", release.TagName, currentVersion)
				fmt.Println("👉 Run 'jar-cart self-update' to pull the latest performance optimizations instantly!\n")
			}
		}
	}()
}

type progressWriter struct {
	total     int64
	current   int64
	lastPrint time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.current += int64(n)
	if time.Since(pw.lastPrint) > 200*time.Millisecond {
		pw.lastPrint = time.Now()
		if pw.total > 0 {
			fmt.Printf("\r📥 Downloading Java: %.2f MB / %.2f MB (%.1f%%)", 
				float64(pw.current)/(1024*1024), 
				float64(pw.total)/(1024*1024), 
				(float64(pw.current)/float64(pw.total))*100)
		} else {
			fmt.Printf("\r📥 Downloading Java: %.2f MB...", float64(pw.current)/(1024*1024))
		}
	}
	return n, nil
}

func EnsureJavaVersion(targetVersion string) error {
	if targetVersion == "" {
		targetVersion = "25"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to locate home directory: %v", err)
	}

	jdkRootPath := filepath.Join(homeDir, ".jar-cart", "jdks")
	targetJdkPath := filepath.Join(jdkRootPath, targetVersion)

	javaExe := "java"
	if runtime.GOOS == "windows" {
		javaExe = "java.exe"
	}

	checkPath := filepath.Join(targetJdkPath, "bin", javaExe)
	if _, err := os.Stat(checkPath); err == nil {
		return nil
	}

	if entries, err := os.ReadDir(targetJdkPath); err == nil && len(entries) > 0 {
		for _, entry := range entries {
			if entry.IsDir() {
				nestedBin := filepath.Join(targetJdkPath, entry.Name(), "bin", javaExe)
				if _, err := os.Stat(nestedBin); err == nil {
					return nil
				}
			}
		}
	}

	fmt.Printf("📦 Java %s is required by this project but not found inside jar-cart context.\n", targetVersion)
	fmt.Printf("⚡ Automatically provisioning isolated Java %s runtime directly from Adoptium...\n", targetVersion)

	osTarget := runtime.GOOS
	if osTarget == "windows" {
		osTarget = "windows"
	} else if osTarget == "darwin" {
		osTarget = "mac"
	} else {
		osTarget = "linux"
	}

	archTarget := "x64"
	if runtime.GOARCH == "arm64" {
		archTarget = "aarch64"
	}

	adoptiumURL := fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%s/ga/%s/%s/jdk/hotspot/normal/eclipse", 
		targetVersion, osTarget, archTarget)

	tempDownloadFile := filepath.Join(os.TempDir(), fmt.Sprintf("java-%s-auto.zip", targetVersion))

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(adoptiumURL)
	if err != nil {
		return fmt.Errorf("failed to reach Adoptium servers: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("adoptium server rejected download request: %s", resp.Status)
	}

	out, err := os.Create(tempDownloadFile)
	if err != nil {
		return err
	}
	defer out.Close()

	pw := &progressWriter{
		total:     resp.ContentLength,
		lastPrint: time.Now(),
	}

	_, err = io.Copy(out, io.TeeReader(resp.Body, pw))
	if err != nil {
		return fmt.Errorf("\ndownload disrupted: %v", err)
	}
	out.Close()
	fmt.Println("\n✨ Download complete. Preparing extraction layers...")

	_ = os.RemoveAll(targetJdkPath)
	_ = os.MkdirAll(targetJdkPath, 0755)

	fmt.Println("📦 Unpacking system components directly into storage pool...")
	if err := unzipJdkArchive(tempDownloadFile, targetJdkPath); err != nil {
		_ = os.RemoveAll(targetJdkPath)
		_ = os.Remove(tempDownloadFile)
		return fmt.Errorf("failed to unpack runtime components safely: %v", err)
	}

	_ = os.Remove(tempDownloadFile)
	fmt.Printf("🎉 Java %s successfully isolated and ready! Continuing execution...\n", targetVersion)
	return nil
}

func unzipJdkArchive(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		parts := strings.Split(f.Name, "/")
		if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
			parts = strings.Split(f.Name, "\\")
		}
		
		if len(parts) <= 1 {
			continue
		}
		
		strippedPath := filepath.Join(parts[1:]...)
		fpath := filepath.Join(dest, strippedPath)

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func SelfUpdate(currentVersion string) error {
	fmt.Println("🔍 Checking GitHub for latest jar-cart CLI release...")
	
	resp, err := http.Get("https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch live metadata: %v", err)
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release layout: %v", err)
	}

	if release.TagName == currentVersion {
		fmt.Printf("✨ You are already on the absolute latest version (%s)!\n", currentVersion)
		return nil
	}

	fmt.Printf("🔄 New version found: %s (Current: %s). Preparing asset downloads...\n", release.TagName, currentVersion)

	var platform, ext string
	switch runtime.GOOS {
	case "windows":
		platform = "windows"
		ext = "zip"
	case "darwin":
		platform = "macos"
		ext = "tar.gz"
	default:
		platform = "linux"
		ext = "tar.gz"
	}

	arch := "x86_64"
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "amd64" {
		if runtime.GOARCH == "arm64" {
			arch = "aarch64"
		}
	}

	fileName := fmt.Sprintf("jar-cart-%s-%s.%s", arch, platform, ext)
	downloadURL := fmt.Sprintf("https://github.com/Sudhanshu-Ambastha/jar-cart/releases/download/%s/%s", release.TagName, fileName)

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate running binary workspace path: %v", err)
	}

	tmpFile := execPath + ".tmp"
	oldFile := execPath + ".old"

	fmt.Println("⚡ Streaming latest binary distribution package payload...")
	out, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary update space: %v", err)
	}
	defer out.Close()

	downloadResp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to pull down update pack: %v", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return fmt.Errorf("download endpoint refused handshake with code: %s", downloadResp.Status)
	}

	_, err = io.Copy(out, downloadResp.Body)
	if err != nil {
		return fmt.Errorf("interrupted download transfer: %v", err)
	}
	out.Close()
	
	_ = os.Remove(oldFile) 
	if err := os.Rename(execPath, oldFile); err != nil {
		return fmt.Errorf("failed to cycle existing binary file handle: %v", err)
	}

	if err := os.Rename(tmpFile, execPath); err != nil {
		_ = os.Rename(oldFile, execPath)
		return fmt.Errorf("failed to lock down update replacement binary: %v", err)
	}

	if runtime.GOOS != "windows" {
		_ = os.Chmod(execPath, 0755)
	}

	fmt.Printf("🎉 Successfully updated jar-cart to %s! Run it again to verify.\n", release.TagName)
	return nil
}

func JdkUpdate(targetVersion string) error {
	if targetVersion == "" {
		targetVersion = "25" 
	}

	fmt.Printf("🔍 Querying Adoptium API endpoints for latest Java %s binary patch layers...\n", targetVersion)

	osTarget := runtime.GOOS
	if osTarget == "darwin" {
		osTarget = "mac"
	}

	archTarget := "x64"
	if runtime.GOARCH == "arm64" {
		archTarget = "aarch64"
	}

	adoptiumURL := fmt.Sprintf("https://api.adoptium.net/v3/binary/latest/%s/ga/%s/%s/jdk/hotspot/normal/eclipse", 
		targetVersion, osTarget, archTarget)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not resolve system home scope: %v", err)
	}

	jdkRootPath := filepath.Join(homeDir, ".jar-cart", "jdks")
	targetJdkPath := filepath.Join(jdkRootPath, targetVersion)
	tempDownloadFile := filepath.Join(os.TempDir(), fmt.Sprintf("java-%s-patch.zip", targetVersion))

	client := &http.Client{}
	resp, err := client.Get(adoptiumURL)
	if err != nil {
		return fmt.Errorf("failed to interface with upstream Adoptium hubs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("adoptium API rejected runtime configuration check: %s", resp.Status)
	}

	fmt.Printf("📥 Streaming Java %s components into local runtime space...\n", targetVersion)
	out, err := os.Create(tempDownloadFile)
	if err != nil {
		return fmt.Errorf("failed to cache package data payload local registers: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("interrupted download session block: %v", err)
	}
	out.Close()

	_ = os.RemoveAll(targetJdkPath)
	_ = os.MkdirAll(targetJdkPath, 0755)

	fmt.Println("📦 Unpacking new isolated JDK archive layers...")
	if err := unzipJdkArchive(tempDownloadFile, targetJdkPath); err != nil {
		_ = os.RemoveAll(targetJdkPath)
		_ = os.Remove(tempDownloadFile)
		return fmt.Errorf("unzip extraction execution failed: %v", err)
	}

	_ = os.Remove(tempDownloadFile)
	fmt.Printf("✨ Java %s environment sync complete! Environment safely isolated inside: %s\n", targetVersion, targetJdkPath)
	return nil
}