package utils

import (
	"archive/zip"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func DetectProjectStrategy() string {
	if FileExists("pom.xml") {
		return "maven"
	}
	if FileExists("build.gradle") || FileExists("build.gradle.kts") {
		return "gradle"
	}
	return "manual"
}

func RunBuild() error {
	strategy := DetectProjectStrategy()
	fmt.Printf("🏗️  Building project using strategy: \033[36m%s\033[0m\n", strategy)

	switch strategy {
	case "maven":
		return runCommand("mvn", "package")
	case "gradle":
		return runCommand("gradle", "build")
	case "manual":
		return performManualBuild()
	default:
		return fmt.Errorf("unknown strategy: %s", strategy)
	}
}

func performManualBuild() error {
	_, javacPath, _, err := GetJDKPaths()
	if err != nil {
		return err
	}
	jarPath := filepath.Join(filepath.Dir(javacPath), "jar")
	if runtime.GOOS == "windows" {
		jarPath += ".exe"
	}

	fmt.Println("⚡ Compiling source tree with managed JDK...")
	_ = os.MkdirAll("bin", 0755)

	var files []string
	filepath.Walk("src", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(path) == ".java" {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		return fmt.Errorf("no .java files found in src/")
	}

	classpath := "bin"
	libFiles, _ := filepath.Glob("lib/*.jar")
	for _, lib := range libFiles {
		classpath += string(os.PathListSeparator) + lib
	}

	err = runCommand(javacPath, append([]string{"-cp", classpath, "-d", "bin"}, files...)...)
	if err != nil {
		return err
	}
	
	fmt.Println("📜 Generating manifest...")
	_ = os.MkdirAll("dist", 0755)
	manifestPath := "dist/manifest.txt"
	manifestContent := "Manifest-Version: 1.0\nMain-Class: App\n"
	err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	fmt.Println("📦 Packaging JAR with managed JDK...")
	return runCommand(jarPath, "cvfm", "dist/app.jar", manifestPath, "-C", "bin", ".")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getMainClassFromJar(jarPath string) (string, error) {
	r, err := zip.OpenReader(jarPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "META-INF/MANIFEST.MF" {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			defer rc.Close()

			scanner := bufio.NewScanner(rc)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "Main-Class: ") {
					return strings.TrimPrefix(line, "Main-Class: "), nil
				}
			}
			break
		}
	}
	return "", fmt.Errorf("Main-Class attribute not found in manifest")
}

func RunJar(jarPath string, mainClass string) error {
	version, _, javaPath, err := GetJDKPaths()
	if err != nil {
		return fmt.Errorf("runtime environment error: %w", err)
	}

	if mainClass == "" {
		detected, err := getMainClassFromJar(jarPath)
		if err != nil {
			fmt.Printf("⚠️  Could not detect main class, falling back to App: %v\n", err)
			mainClass = "App"
		} else {
			mainClass = detected
		}
	}

	libFiles, _ := filepath.Glob("lib/*.jar")
	classpath := jarPath
	for _, lib := range libFiles {
		classpath += string(os.PathListSeparator) + lib
	}
	args := []string{
		"--enable-native-access=ALL-UNNAMED",
		"-cp", classpath,
		mainClass,
	}

	fmt.Printf("🚀 Launching with JDK %s [Main: %s]\n", version, mainClass)
	return runCommand(javaPath, args...)
}