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

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
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

func DetectModuleStrategy(modulePath string) string {
	if FileExists(filepath.Join(modulePath, "pom.xml")) {
		return "maven"
	}
	if FileExists(filepath.Join(modulePath, "build.gradle")) || FileExists(filepath.Join(modulePath, "build.gradle.kts")) {
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

func RunWorkspaceBuild(workspace models.WorkspaceManifest) error {
	log.Info("Starting workspace-wide multi-module build...")

	for modName, modConfig := range workspace.Modules {
		log.Info("Building workspace module", "module", modName, "path", modConfig.Path)
		
		if _, err := os.Stat(modConfig.Path); os.IsNotExist(err) {
			log.Warn("Module directory does not exist, skipping", "module", modName, "path", modConfig.Path)
			continue
		}

		originalDir, err := os.Getwd()
		if err != nil {
			return err
		}

		if err := os.Chdir(modConfig.Path); err != nil {
			return fmt.Errorf("failed to enter module directory %s: %w", modConfig.Path, err)
		}

		strategy := DetectModuleStrategy(".")
		var buildErr error

		switch strategy {
		case "maven":
			buildErr = runCommand("mvn", "package")
		case "gradle":
			buildErr = runCommand("gradle", "build")
		case "manual":
			buildErr = performManualBuild()
		default:
			buildErr = fmt.Errorf("unknown strategy for module %s", modName)
		}

		_ = os.Chdir(originalDir)

		if buildErr != nil {
			return fmt.Errorf("failed building module %s: %w", modName, buildErr)
		}
	}

	log.Info("All workspace modules built successfully!")
	return nil
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

func OptimizeJarForDeployment(jarPath string, outputDir string) error {
	if !FileExists(jarPath) {
		return fmt.Errorf("JAR file not found: %s. Did you forget to run 'jar-cart build'?", jarPath)
	}

	initialInfo, err := os.Stat(jarPath)
	if err != nil {
		return fmt.Errorf("failed to read source JAR info: %w", err)
	}
	initialSize := initialInfo.Size()

	if _, err := os.Stat(outputDir); err == nil {
		log.Info("Cleaning existing output directory...", "path", outputDir)
		if err := os.RemoveAll(outputDir); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	optimizer := &JLinkOptimizer{JarPath: jarPath}
	javaVersion := "21"
	cfg := models.OptimizationConfig{
		Compression: 2,
		StripDebug:  true,
		StripNative: false,
	}

	manifestFiles := []string{"jar-cart.json", "jar-cart.xml"}
	for _, file := range manifestFiles {
		if FileExists(file) {
			adapter := GetAdapterForFile(file)
			if adapter != nil {
				manifest, err := adapter.Load(file)
				if err == nil {
					if manifest.JavaVersion != "" {
						javaVersion = manifest.JavaVersion
					}
					if manifest.Optimize.Compression != 0 {
						cfg.Compression = manifest.Optimize.Compression
					}
					cfg.StripDebug = manifest.Optimize.StripDebug
					cfg.StripNative = manifest.Optimize.StripNative
					break
				}
			}
		}
	}
	
	log.Info("Analyzing dependencies with jdeps...", "version", javaVersion)
	modules, err := optimizer.AnalyzeDependencies(javaVersion)
	if err != nil {
		return err
	}
	
	log.Info("Trimming runtime with jlink...", "modules", modules, "compression", cfg.Compression)
	if err := optimizer.CreateCustomRuntime(modules, outputDir, cfg); err != nil {
		return err
	}

	finalSize, err := getDirectorySize(outputDir)
	if err != nil {
		log.Warn("Could not compute final runtime directory size", "error", err)
		log.Info("Runtime optimized successfully!")
		return nil
	}

	diffBytes := initialSize - finalSize
	var percentage float64
	if initialSize > 0 {
		percentage = (float64(diffBytes) / float64(initialSize)) * 100
	}

	log.Info("Optimization Metrics", 
		"source_jar_size", formatBytes(initialSize), 
		"optimized_runtime_size", formatBytes(finalSize), 
		"net_reduction", fmt.Sprintf("%.2f%%", percentage),
	)
	log.Info("Runtime optimized successfully!")
	return nil
}

func RunJar(target string, mainClass string, appArgs []string) error {
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

	log.Info(
		"Launching JAR",
		"path", jarPath,
		"jdk", version,
		"main", mainClass,
		"args", appArgs,
	)

	args := []string{
		"-cp",
		jarPath,
		mainClass,
	}

	args = append(args, appArgs...)

	cmd := exec.Command(javaPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}