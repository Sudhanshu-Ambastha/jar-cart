package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

var Verbose bool

type SkeletonStrategy struct {
	Directories []string          `json:"directories"`
	Files       map[string]string `json:"files"`
}

type RemoteRegistry struct {
	Strategies map[string]SkeletonStrategy `json:"strategies"`
}

type ProjectData struct {
	Name  string
	Group string
}

type Config struct {
	Project      string            `json:"project"`
	Strategy     string            `json:"strategy"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies []interface{}     `json:"dependencies"`
}

func logDebug(format string, a ...interface{}) {
	if Verbose {
		fmt.Printf("DEBUG: "+format+"\n", a...)
	}
}

func getRegistryURL() string {
	if envURL := os.Getenv("JAR_CART_REGISTRY_URL"); envURL != "" {
		return envURL
	}
	return "https://raw.githubusercontent.com/Sudhanshu-Ambastha/jar-cart/main/registry.json"
}

func CleanCache() error {
	home, _ := os.UserHomeDir()
	return os.RemoveAll(filepath.Join(home, ".jar-cart", "cache"))
}

func getLocalRegistryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".jar-cart", "registry.json")
}

func fetchAndCacheRegistry() RemoteRegistry {
	cacheFile := getLocalRegistryPath()
	os.MkdirAll(filepath.Dir(cacheFile), 0755)

	info, err := os.Stat(cacheFile)
	if err != nil || time.Since(info.ModTime()) > 24*time.Hour {
		logDebug("Fetching fresh strategy registry from remote: %s", getRegistryURL())
		req, _ := http.NewRequest("GET", getRegistryURL(), nil)
		if resp, err := http.DefaultClient.Do(req); err == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				if out, err := os.Create(cacheFile + ".tmp"); err == nil {
					_, _ = io.Copy(out, resp.Body)
					out.Close()
					_ = os.Rename(cacheFile+".tmp", cacheFile)
				}
			}
		}
	}

	content, err := os.ReadFile(cacheFile)
	var reg RemoteRegistry
	if err == nil {
		_ = json.Unmarshal(content, &reg)
	}

	if reg.Strategies == nil {
		reg.Strategies = getDefaultFallbackStrategies()
	}
	return reg
}

func getDefaultFallbackStrategies() map[string]SkeletonStrategy {
	return map[string]SkeletonStrategy{
		"flat": {
			Directories: []string{"src"},
			Files: map[string]string{
				"src/App.java": "package src;\n\npublic class App {\n    public static void main(String[] args) {\n        System.out.println(\"Hello from jar-cart! 🚀\");\n    }\n}",
			},
		},
		"backend": {
			Directories: []string{
				"src/com/srs/db",
				"src/db",
				"src/resources",
				"src/sql",
				"lib",
				"bin",
			},
			Files: map[string]string{
				"src/com/srs/App.java": `package com.srs;

public class App {
    public static void main(String[] args) {
        System.out.println("Backend application initialized successfully! 🏢");
    }
}`,
				"src/com/srs/db/DBconnections.java": `package com.srs.db;

public class DBconnections {
    // TODO: Implement database connection management
}`,
				".env": `PORT=8080
DB_PATH=src/db/srs.db`,
			},
		},
	}
}

func HandleInit(projectName, manifestType, javaVersion, strategy string) (string, error) {
	var targetDir string
	if projectName == "." {
		targetDir, _ = os.Getwd()
	} else {
		targetDir = projectName
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", err
		}
	}

	if strategy == "" {
		strategy = "flat"
	}

	registry := fetchAndCacheRegistry()
	strat, exists := registry.Strategies[strategy]
	if !exists {
		strat = registry.Strategies["flat"]
	}

	for _, subDir := range strat.Directories {
		fullSub := targetDir
		if subDir != "" {
			fullSub = filepath.Join(targetDir, subDir)
		}
		_ = os.MkdirAll(fullSub, 0755)
	}

	for relPath, content := range strat.Files {
		fullFilePath := filepath.Join(targetDir, relPath)
		_ = os.MkdirAll(filepath.Dir(fullFilePath), 0755)
		_ = os.WriteFile(fullFilePath, []byte(content), 0644)
	}

	jsonPath := filepath.Join(targetDir, "jar-cart.json")
	xmlPath := filepath.Join(targetDir, "jar-cart.xml")
	if manifestType == "xml" {
		_ = os.Remove(jsonPath)
		xmlContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Manifest>
    <project>%s</project>
    <java_version>%s</java_version>
    <strategy>%s</strategy>
    <scripts>
        <script name="hello">echo 'Hello from jar-cart!'</script>
        <script name="pretest">echo 'Compiling tests...'</script>
        <script name="test">echo 'Running tests...'</script>
        <script name="posttest">echo 'Cleaning up test artifacts...'</script>
    </scripts>
    <dependencies></dependencies>
</Manifest>`, filepath.Base(targetDir), javaVersion, strategy)
		_ = os.WriteFile(xmlPath, []byte(xmlContent), 0644)
	} else {
		_ = os.Remove(xmlPath)
		config := models.Manifest{
			Project:         filepath.Base(targetDir),
			JavaVersion:     javaVersion,
			ResolutionDepth: "full",
			Scripts: map[string]string{
				"hello":    "echo 'Hello from jar-cart!'",
				"pretest":  "echo 'Compiling tests...'",
				"test":     "echo 'Running tests...'",
				"posttest": "echo 'Cleaning up test artifacts...'",
			},
			Dependencies: []models.Dependency{},
		}
		if data, err := json.MarshalIndent(config, "", "    "); err == nil {
			_ = os.WriteFile(jsonPath, data, 0644)
		}
	}

	hashFiles := GetSourceFiles(targetDir)
	if hash, err := CalculateProjectHash(hashFiles); err == nil {
		cartDir := filepath.Join(targetDir, ".jar-cart")
		_ = os.MkdirAll(cartDir, 0755)
		_ = os.WriteFile(filepath.Join(cartDir, "last_build.hash"), hash[:], 0644)
	}

	return targetDir, nil
}

func RegisterCustomTemplate(templateKey string, projectPath string) error {
	registry := fetchAndCacheRegistry()
	dirs := []string{}
	files := make(map[string]string)

	absRoot, _ := filepath.Abs(projectPath)
	_ = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(absRoot, path)
		if rel == "." || strings.HasPrefix(rel, ".jar-cart") || strings.HasPrefix(rel, "bin") || strings.HasPrefix(rel, "lib") {
			return nil
		}
		if info.IsDir() {
			dirs = append(dirs, rel)
		} else {
			baseName := filepath.Base(rel)
			if baseName == "jar-cart.json" || baseName == "jar-cart.xml" || baseName == "jar-cart.lock" || baseName == "LICENSE" {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(rel))
			if ext == ".java" || ext == ".properties" || ext == ".env" || ext == ".yml" || ext == ".yaml" || ext == ".sql" {
				if content, err := os.ReadFile(path); err == nil {
					files[rel] = string(content)
				}
			}
		}
		return nil
	})

	registry.Strategies[templateKey] = SkeletonStrategy{
		Directories: dirs,
		Files:       files,
	}

	localPath := getLocalRegistryPath()
	if data, err := json.MarshalIndent(registry, "", "    "); err == nil {
		return os.WriteFile(localPath, data, 0644)
	}
	return nil
}

func ListCustomTemplates() error {
	registry := fetchAndCacheRegistry()
	
	fmt.Println("\n📦 Available Templates in Local Registry:")
	fmt.Println("----------------------------------------")
	
	if len(registry.Strategies) == 0 {
		fmt.Println("  (No templates registered)")
		return nil
	}

	for key, strategy := range registry.Strategies {
		fmt.Printf("🔹 Key: %s\n", key)
		fmt.Printf("   ├── Directories: %d tracked\n", len(strategy.Directories))
		fmt.Printf("   └── Files:       %d tracked\n", len(strategy.Files))
		fmt.Println()
	}
	return nil
}

func RemoveCustomTemplate(templateKey string) error {
	registry := fetchAndCacheRegistry()

	if _, exists := registry.Strategies[templateKey]; !exists {
		return fmt.Errorf("template '%s' not found in registry", templateKey)
	}

	delete(registry.Strategies, templateKey)

	localPath := getLocalRegistryPath()
	if data, err := json.MarshalIndent(registry, "", "    "); err == nil {
		return os.WriteFile(localPath, data, 0644)
	}
	return nil
}

func ExecuteScaffold(projectDir, projectName, framework, strategy, lang, javaVersion, manifestType string) error {
	if javaVersion == "" {
		javaVersion = "25"
	}

	manifestName := "jar-cart.json"
	if manifestType == "xml" {
		manifestName = "jar-cart.xml"
	}
	manifestPath := filepath.Join(projectDir, manifestName)
	srcPath := filepath.Join(projectDir, "src")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) && manifestType == "json" {
		m := models.Manifest{
			Project:         projectName,
			JavaVersion:     javaVersion,
			ResolutionDepth: "full",
			Scripts: map[string]string{
				"hello":    "echo 'Hello from jar-cart!'",
				"pretest":  "echo 'Compiling tests...'",
				"test":     "echo 'Running tests...'",
				"posttest": "echo 'Cleaning up test artifacts...'",
			},
			Dependencies: []models.Dependency{},
		}
		data, _ := json.MarshalIndent(m, "", "    ")
		_ = os.WriteFile(manifestPath, data, 0644)
	}

	os.MkdirAll(srcPath, 0755)

	if strategy == "no-build" {
		return nil
	}

	urlProjectName := strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))
	targetURL, _ := resolveBlueprintURL(framework, strategy, ProjectData{Name: urlProjectName, Group: "src"})

	if targetURL != "" {
		logDebug("Pulling from: %s", targetURL)
		tmpZip := filepath.Join(os.TempDir(), projectName+".zip.tmp")
		finalZip := filepath.Join(os.TempDir(), projectName+".zip")

		if err := downloadFileAtomic(targetURL, tmpZip, finalZip); err == nil {
			return unzipStrippingRoot(finalZip, projectDir)
		}
	}
	return nil
}

func resolveBlueprintURL(framework, strategy string, data ProjectData) (string, error) {
	home, _ := os.UserHomeDir()
	cacheFile := filepath.Join(home, ".jar-cart", "registry.json")
	os.MkdirAll(filepath.Dir(cacheFile), 0755)

	info, err := os.Stat(cacheFile)
	if err != nil || time.Since(info.ModTime()) > 24*time.Hour {
		logDebug("Registry cache expired. Fetching: %s", getRegistryURL())
		req, _ := http.NewRequest("GET", getRegistryURL(), nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			out, _ := os.Create(cacheFile + ".tmp")
			io.Copy(out, resp.Body)
			out.Close()
			os.Rename(cacheFile+".tmp", cacheFile)
		}
	}

	content, _ := os.ReadFile(cacheFile)
	var registry struct {
		Frameworks map[string]map[string]string `json:"frameworks"`
	}
	json.Unmarshal(content, &registry)

	tmplStr := registry.Frameworks[framework][strategy]
	t, _ := template.New("url").Parse(tmplStr)
	var buf bytes.Buffer
	t.Execute(&buf, data)
	return buf.String(), nil
}

func downloadFileAtomic(url, tmpPath, finalPath string) error {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := os.Create(tmpPath)
	io.Copy(out, resp.Body)
	out.Close()
	return os.Rename(tmpPath, finalPath)
}

func unzipStrippingRoot(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(path), 0755)
		outFile, _ := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		rc, _ := f.Open()
		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}
	return nil
}