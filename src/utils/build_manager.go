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

	"github.com/charmbracelet/log"
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
	log.Info("Building project", "strategy", strategy)

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

func detectMainClassInBin(binDir string) (string, error) {
	var foundClass string
	err := filepath.Walk(binDir, func(path string, info os.FileInfo, _ error) error {
		if !info.IsDir() && strings.HasSuffix(path, ".class") {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil 
			}
			
			if strings.Contains(string(data), "main") && strings.Contains(string(data), "PrintStream") {
				rel, _ := filepath.Rel(binDir, path)
				foundClass = strings.ReplaceAll(strings.TrimSuffix(rel, ".class"), string(os.PathSeparator), ".")
			}
		}
		return nil
	})
	
	if err != nil {
		return "", err
	}
	
	if foundClass == "" {
		return "", fmt.Errorf("no main method found")
	}
	return foundClass, nil
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

	log.Info("Compiling source tree")
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

	err = runCommand(javacPath, append([]string{"-cp", "lib/*", "-d", "bin"}, files...)...)
	if err != nil {
		return err
	}

	mainClass, err := detectMainClassInBin("bin")
	if err != nil {
		return fmt.Errorf("could not auto-detect Main-Class: %w", err)
	}
	log.Info("Detected Main-Class", "class", mainClass)

	log.Info("Generating manifest")
	_ = os.MkdirAll("dist", 0755)
	manifestPath := "dist/manifest.txt"
	manifestContent := fmt.Sprintf("Manifest-Version: 1.0\nMain-Class: %s\n", mainClass)
	os.WriteFile(manifestPath, []byte(manifestContent), 0644)

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

func RunJar(target string, mainClass string) error {
	version, _, javaPath, err := GetJDKPaths()
	if err != nil {
		return fmt.Errorf("runtime environment error: %w", err)
	}

	jarPath := target
	info, err := os.Stat(target)
	if os.IsNotExist(err) {
		commonDirs := []string{"dist", "."}
		for _, dir := range commonDirs {
			potential := filepath.Join(dir, target)
			if _, err := os.Stat(potential); err == nil {
				jarPath = potential
				break
			}
			if !strings.HasSuffix(target, ".jar") {
				potential = filepath.Join(dir, target+".jar")
				if _, err := os.Stat(potential); err == nil {
					jarPath = potential
					break
				}
			}
		}
	} else if info.IsDir() {
		files, _ := filepath.Glob(filepath.Join(target, "*.jar"))
		if len(files) == 1 {
			jarPath = files[0]
		} else {
			return fmt.Errorf("directory %s contains multiple or no JARs; please specify one", target)
		}
	}

	if mainClass == "" {
		detected, err := getMainClassFromJar(jarPath)
		if err != nil {
			return fmt.Errorf("could not detect Main-Class in %s: %w", jarPath, err)
		}
		mainClass = detected
	}

	log.Info("Launching JAR", "path", jarPath, "jdk", version, "main", mainClass)
	return runCommand(javaPath, "-cp", jarPath, mainClass)
}