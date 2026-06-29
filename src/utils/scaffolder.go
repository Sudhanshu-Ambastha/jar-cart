package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

var Verbose bool

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

func HandleInit(projectName, manifestType string) (string, error) {
	var targetDir string

	if projectName == "." {
		targetDir, _ = os.Getwd()
	} else {
		targetDir = projectName
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", err
		}
	}

	jsonPath := filepath.Join(targetDir, "jar-cart.json")
	xmlPath := filepath.Join(targetDir, "jar-cart.xml")

	if manifestType == "xml" {
		_ = os.Remove(jsonPath)
	} else {
		_ = os.Remove(xmlPath)
	}

	for _, dir := range []string{"bin", "lib", "src"} {
		if err := os.MkdirAll(filepath.Join(targetDir, dir), 0755); err != nil {
			return targetDir, err
		}
	}
	
	appCode := `package src;

public class App {
    public static void main(String[] args) {
        System.out.println("Hello, jar-cart! Your project is ready. 🚀");
    }
}`
	if err := os.WriteFile(filepath.Join(targetDir, "src", "App.java"), []byte(appCode), 0644); err != nil {
		return targetDir, err
	}

	javaVersion := "25"

	var err error
	if manifestType == "xml" {
		xmlContent := `<?xml version="1.0" encoding="UTF-8"?>
<Manifest>
    <project>` + filepath.Base(targetDir) + `</project>
    <java_version>` + javaVersion + `</java_version>
    <strategy>Include All Dependencies</strategy>
    <scripts>
        <script name="hello">echo 'Hello from jar-cart!'</script>
        <script name="pretest">echo 'Compiling tests...'</script>
        <script name="test">echo 'Running tests...'</script>
        <script name="posttest">echo 'Cleaning up test artifacts...'</script>
    </scripts>
    <dependencies></dependencies>
</Manifest>`
		err = os.WriteFile(xmlPath, []byte(xmlContent), 0644)
	} else {
		config := models.Manifest{
			Project:     filepath.Base(targetDir),
			JavaVersion: javaVersion,
			Strategy:    "Include All Dependencies",
			Scripts: map[string]string{
				"hello":    "echo 'Hello from jar-cart!'",
				"pretest":  "echo 'Compiling tests...'",
				"test":     "echo 'Running tests...'",
				"posttest": "echo 'Cleaning up test artifacts...'",
			},
			Dependencies: []models.Dependency{},
		}
		
		var configData []byte
		configData, err = json.MarshalIndent(config, "", "    ")
		if err == nil {
			err = os.WriteFile(jsonPath, configData, 0644)
		}
	}
	
	return targetDir, err
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
			Project:     projectName,
			JavaVersion: javaVersion,
			Strategy:    strategy,
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
	code := `package src;

public class App { 
    public static void main(String[] args) { 
        System.out.println("Hello from jar-cart!"); 
    } 
}`
	_ = os.WriteFile(filepath.Join(srcPath, "App.java"), []byte(code), 0644)
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

func generateFallbackTemplate(projectDir, projectName, strategy string) error {
	srcPath := filepath.Join(projectDir, "src")
	os.MkdirAll(srcPath, 0755)
	code := `package src;

public class App { 
    public static void main(String[] args) { 
        System.out.println("Hello from jar-cart!"); 
    } 
}`
	return os.WriteFile(filepath.Join(srcPath, "App.java"), []byte(code), 0644)
}

func LaunchWorkspace(baseDir string) {
	cmd := exec.Command("code", baseDir)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "code", baseDir)
	}
	cmd.Start()
}