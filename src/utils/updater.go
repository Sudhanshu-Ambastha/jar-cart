package utils

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/mod/semver"
)

const MinSupportedVersion = "v0.2.1"

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type UpdateCache struct {
    LatestVersion string    `json:"latest_version"`
    CheckedAt     time.Time `json:"checked_at"`
    ETag          string    `json:"etag"` 
}

type progressWriter struct {
	total     int64
	current   int64
	lastPrint time.Time
}

func normalizeVersion(version string) string {
	if version == "" {
		return ""
	}

	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	return version
}

func ValidateTargetVersion(version string) error {
	version = normalizeVersion(version)

	if !semver.IsValid(version) {
		return fmt.Errorf("invalid version: %s", version)
	}

	if semver.Compare(version, MinSupportedVersion) < 0 {
		return fmt.Errorf(
			"versions older than %s are no longer supported",
			MinSupportedVersion,
		)
	}

	return nil
}

func unzipBinary(src, destDir string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".exe") || f.Name == "jar-cart" {
			outPath := filepath.Join(destDir, "jar-cart.exe")
			if runtime.GOOS != "windows" {
				outPath = filepath.Join(destDir, "jar-cart")
			}
			
			outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
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
			return err
		}
	}
	return fmt.Errorf("no executable found in archive")
}

func saveLatestVersionCache(version string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	cacheDir := filepath.Join(homeDir, ".jar-cart")
	_ = os.MkdirAll(cacheDir, 0755)

	cache := UpdateCache{
		LatestVersion: version,
		CheckedAt:     time.Now(),
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(filepath.Join(cacheDir, "latest_version.json"), data, 0644)
}

func loadLatestVersionCache() (*UpdateCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(homeDir, ".jar-cart", "latest_version.json"))
	if err != nil {
		return nil, err
	}

	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func saveLatestVersionCacheWithETag(version, etag string) {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".jar-cart")
	cache := UpdateCache{
		LatestVersion: version,
		CheckedAt:     time.Now(),
		ETag:          etag,
	}
	data, _ := json.MarshalIndent(cache, "", "  ")
	_ = os.WriteFile(filepath.Join(cacheDir, "latest_version.json"), data, 0644)
}

func AutoCheckUpdate(currentVersion string) (bool, string) {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".jar-cart")
	timestampFile := filepath.Join(cacheDir, ".last_update_check")
	if info, err := os.Stat(timestampFile); err != nil || time.Since(info.ModTime()) >= 30*time.Minute {
		_ = os.WriteFile(timestampFile, []byte(time.Now().Format(time.RFC3339)), 0644)

		go func() {
			cache, _ := loadLatestVersionCache()
			etag := ""
			if cache != nil { etag = cache.ETag }

			release, newETag, err := FetchReleaseMetadataWithETag("latest", etag)
			if err == nil && newETag != "" && newETag != etag {
				saveLatestVersionCacheWithETag(release.TagName, newETag)
			}
		}()
	}

	cache, _ := loadLatestVersionCache()
	if cache == nil || cache.LatestVersion == "" { return false, "" }

	if semver.Compare(normalizeVersion(currentVersion), normalizeVersion(cache.LatestVersion)) < 0 {
		return true, cache.LatestVersion
	}
	return false, ""
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.current += int64(n)
	if time.Since(pw.lastPrint) > 200*time.Millisecond {
		pw.lastPrint = time.Now()
		if pw.total > 0 {
			fmt.Printf("\r📥 Downloading JAR: %.2f MB / %.2f MB (%.1f%%)", 
				float64(pw.current)/(1024*1024), 
				float64(pw.total)/(1024*1024), 
				(float64(pw.current)/float64(pw.total))*100)
		} else {
			fmt.Printf("\r📥 Downloading JAR: %.2f MB...", float64(pw.current)/(1024*1024))
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

func FetchReleaseMetadataWithETag(tag string, currentETag string) (GitHubRelease, string, error) {
    url := "https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/latest"
    if tag != "latest" {
        url = fmt.Sprintf("https://api.github.com/repos/Sudhanshu-Ambastha/jar-cart/releases/tags/%s", tag)
    }
    
    req, _ := http.NewRequest("GET", url, nil)
    if currentETag != "" {
        req.Header.Set("If-None-Match", currentETag)
    }

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil { 
        return GitHubRelease{}, "", err 
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotModified {
        return GitHubRelease{}, currentETag, nil 
    }

    var release GitHubRelease
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return GitHubRelease{}, "", err
    }
    
    return release, resp.Header.Get("ETag"), nil
}

func FetchReleaseMetadata(tag string) (GitHubRelease, error) {
    release, _, err := FetchReleaseMetadataWithETag(tag, "")
    return release, err
}

func DownloadAndVerify(release GitHubRelease, tmpFile string) error {
	platform, arch := runtime.GOOS, "x86_64"
	if runtime.GOARCH == "arm64" { arch = "aarch64" }
	ext := "tar.gz"
	if platform == "windows" { ext = "zip" }

	fileName := fmt.Sprintf("jar-cart-%s-%s.%s", arch, platform, ext)
	var downloadURL, expectedHash string
	for _, asset := range release.Assets {
		if asset.Name == fileName {
			downloadURL = asset.BrowserDownloadURL
			expectedHash = strings.TrimPrefix(asset.Digest, "sha256:")
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no matching asset found for %s", fileName)
	}

	resp, err := http.Get(downloadURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed")
	}
	defer resp.Body.Close()

	out, err := os.Create(tmpFile)
	if err != nil { return err }
	defer out.Close()
	
	if _, err = io.Copy(out, resp.Body); err != nil { return err }

	f, _ := os.Open(tmpFile)
	defer f.Close()
	h := sha256.New()
	io.Copy(h, f)

	if hex.EncodeToString(h.Sum(nil)) != strings.ToLower(expectedHash) {
		return fmt.Errorf("security alert: checksum mismatch")
	}
	return nil
}

func ApplyBinarySwap(tmpFile string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	oldFile := execPath + ".old"

	if runtime.GOOS == "windows" {
		extractDir := filepath.Join(os.TempDir(), "jar-cart-update")
		_ = os.RemoveAll(extractDir)

		if err := os.MkdirAll(extractDir, 0755); err != nil {
			return fmt.Errorf("failed to create extraction directory: %w", err)
		}
		defer os.RemoveAll(extractDir)
		if err := unzipBinary(tmpFile, extractDir); err != nil {
			return err
		}

		newExe := filepath.Join(extractDir, "jar-cart.exe")
		_ = os.Remove(oldFile)

		if err := os.Rename(execPath, oldFile); err != nil {
			return fmt.Errorf("failed to backup current executable: %w", err)
		}

		if err := os.Rename(newExe, execPath); err != nil {
			cmd := exec.Command("cmd", "/C",
				fmt.Sprintf(`move /Y "%s" "%s"`, newExe, execPath))

			if err := cmd.Run(); err != nil {
				_ = os.Rename(oldFile, execPath)
				return fmt.Errorf("failed to replace executable: %w", err)
			}
		}

		cleanup := exec.Command(
			"cmd",
			"/C",
			fmt.Sprintf(`ping 127.0.0.1 -n 3 > nul && del /F /Q "%s"`, oldFile),
		)

		_ = cleanup.Start()

	} else {
		_ = os.Remove(oldFile)

		if err := os.Rename(execPath, oldFile); err != nil {
			return fmt.Errorf("failed to backup executable: %w", err)
		}

		if err := os.Rename(tmpFile, execPath); err != nil {
			_ = os.Rename(oldFile, execPath)
			return fmt.Errorf("failed to install updated executable: %w", err)
		}

		if err := os.Chmod(execPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}

		if err := os.Remove(oldFile); err != nil {
			log.Warn("Could not remove backup executable", "file", oldFile, "error", err)
		}
	}

	return nil
}

func PerformUpdate(tag string) error {
    release, _, err := FetchReleaseMetadataWithETag(tag, "") 
    if err != nil { return err }
    
    tmpFile := filepath.Join(os.TempDir(), "jar-cart.tmp")
    defer os.Remove(tmpFile)

    if err := DownloadAndVerify(release, tmpFile); err != nil { return err }
    return ApplyBinarySwap(tmpFile)
}

func SelfUpdate(currentVersion string) error {
    logger := log.New(os.Stderr)
    logger.Info("Checking GitHub for latest release...")
    release, err := FetchReleaseMetadata("latest") 
    if err != nil {
        return err
    }

    c := normalizeVersion(currentVersion)
    l := normalizeVersion(release.TagName)

    if semver.Compare(c, l) >= 0 {
        logger.Info("Already on the latest version", "version", currentVersion)
        return nil
    }

    logger.Info("Updating to new version", "version", release.TagName)
    return PerformUpdate(release.TagName)
}

func DowngradeTo(targetVersion string) error {
    if err := ValidateTargetVersion(targetVersion); err != nil {
        return err
    }

    return PerformUpdate(targetVersion)
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