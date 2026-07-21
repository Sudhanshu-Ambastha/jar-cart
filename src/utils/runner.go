package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/charmbracelet/log"
)

func FindJavaFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".java") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func getInstalledJavaVersion(binPath string) (int, error) {
	cmd := exec.Command(filepath.Join(binPath, "java"), "-version")
	output, _ := cmd.CombinedOutput() 
	outStr := string(output)
	if strings.Contains(outStr, "version \"17") { return 17, nil }
	if strings.Contains(outStr, "version \"25") { return 25, nil }
	return 0, fmt.Errorf("could not detect version")
}

func ResolveMainClass(filePath string) string {
    content, err := os.ReadFile(filePath)
    if err != nil {
        return strings.TrimSuffix(filepath.Base(filePath), ".java")
    }

    lines := strings.Split(string(content), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "package ") {
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                pkg := strings.TrimSuffix(parts[1], ";")
                className := strings.TrimSuffix(filepath.Base(filePath), ".java")
                return pkg + "." + className
            }
        }
    }
    return strings.TrimSuffix(filepath.Base(filePath), ".java")
}

func executeProject(dir string, input string, appArgs []string, workspace *models.WorkspaceManifest, isModule bool) {
	javaVersion, javacPath, javaPath, err := GetJDKPaths()
	if err != nil {
		log.Error("JDK configuration error", "error", err)
		return
	}

	javaFiles, err := FindJavaFiles(dir)
	if err != nil || len(javaFiles) == 0 {
		log.Error("No Java source files detected", "path", dir)
		return
	}

	binDir, _ := filepath.Abs(filepath.Join(dir, ".jar-cart", "bin"))
	hashFilePath := filepath.Join(dir, ".jar-cart", "last_build.hash")
	absLib, _ := filepath.Abs(filepath.Join(dir, "lib"))

	classpath := binDir + string(os.PathListSeparator) + filepath.Join(absLib, "*")
	if isModule && workspace != nil {
		workspaceCp := BuildWorkspaceClasspath(*workspace, dir)
		if workspaceCp != "" {
			classpath = workspaceCp + string(os.PathListSeparator) + classpath
		}
	}

	newHash, _ := CalculateProjectHash(javaFiles)
	lastHashBytes, _ := os.ReadFile(hashFilePath)

	if string(lastHashBytes) != fmt.Sprintf("%x", newHash) {
		log.Info("Changes detected, recompiling...", "path", dir, "version", javaVersion)

		_ = os.RemoveAll(binDir)
		_ = os.MkdirAll(filepath.Join(dir, ".jar-cart"), 0755)
		_ = os.MkdirAll(binDir, 0755)

		argfilePath := filepath.Join(dir, ".jar-cart", "sources.txt")
		argfile, err := os.Create(argfilePath)
		if err != nil {
			log.Error("Failed to create javac argfile", "error", err)
			return
		}
		for _, file := range javaFiles {
			relPath, _ := filepath.Rel(dir, file)
			_, _ = argfile.WriteString(filepath.ToSlash(relPath) + "\n")
		}
		argfile.Close()
		defer os.Remove(argfilePath)

		absArgFilePath, _ := filepath.Abs(argfilePath)
		javacCmd := exec.Command(javacPath, "-cp", classpath, "-sourcepath", dir, "-d", binDir, "@"+absArgFilePath)
		javacCmd.Dir = dir
		javacCmd.Stdout, javacCmd.Stderr = os.Stdout, os.Stderr

		if err := javacCmd.Run(); err != nil {
			log.Error("Compilation failed", "path", dir, "error", err)
			return
		}
		_ = os.WriteFile(hashFilePath, []byte(fmt.Sprintf("%x", newHash)), 0644)
	} else {
		log.Info("No changes detected, skipping compilation.", "path", dir)
	}

	if isModule && input == "" {
		var mainFile string
		for _, f := range javaFiles {
			if hasMainMethod(f) {
				mainFile = f
				break
			}
		}
		if mainFile == "" {
			log.Error("No main method found in module source files", "path", dir)
			return
		}
		targetFile := mainFile
		mainClass := ResolveMainClass(targetFile)
		log.Info("Booting execution engine", "path", dir, "class", mainClass)

		args := append([]string{"-cp", classpath, mainClass}, appArgs...)
		javaCmd := exec.Command(javaPath, args...)
		javaCmd.Stdin, javaCmd.Stdout, javaCmd.Stderr = os.Stdin, os.Stdout, os.Stderr

		if err := javaCmd.Run(); err != nil {
			log.Error("Execution failed", "path", dir, "error", err)
		}
		return
	}

	targetFile, err := resolveTargetFile(dir, input, javaFiles)
	if err != nil {
		log.Error("Could not resolve target", "path", dir, "input", input, "error", err)
		return
	}

	mainClass := ResolveMainClass(targetFile)
	log.Info("Booting execution engine", "path", dir, "class", mainClass)

	args := append([]string{"-cp", classpath, mainClass}, appArgs...)
	javaCmd := exec.Command(javaPath, args...)
	javaCmd.Stdin, javaCmd.Stdout, javaCmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := javaCmd.Run(); err != nil {
		log.Error("Execution failed", "path", dir, "error", err)
	}
}

func RunProject(input string, appArgs []string) {
	cwd, _ := os.Getwd()
	executeProject(cwd, input, appArgs, nil, false)
}

func RunModuleProject(modulePath string, input string, appArgs []string, workspace models.WorkspaceManifest) {
	executeProject(modulePath, input, appArgs, &workspace, true)
}

func resolveTargetFile(cwd, input string, javaFiles []string) (string, error) {
	cleanInput := strings.TrimSuffix(input, ".java")

	candidate := input
	if !strings.HasSuffix(candidate, ".java") {
		candidate += ".java"
	}
	if absCandidate, err := filepath.Abs(candidate); err == nil {
		for _, f := range javaFiles {
			if f == absCandidate {
				return f, nil
			}
		}
	}

	for _, f := range javaFiles {
		base := strings.TrimSuffix(filepath.Base(f), ".java")
		if base == cleanInput {
			return f, nil
		}
	}

	dirCandidate := filepath.Join(cwd, input)
	if info, err := os.Stat(dirCandidate); err == nil && info.IsDir() {
		for _, f := range javaFiles {
			if strings.HasPrefix(f, dirCandidate) {
				if hasMainMethod(f) {
					return f, nil
				}
			}
		}
	}

	var mains []string
	for _, f := range javaFiles {
		if hasMainMethod(f) {
			mains = append(mains, f)
		}
	}
	if len(mains) == 1 {
		return mains[0], nil
	}

	return "", fmt.Errorf("no matching Java file found for '%s'", input)
}

func hasMainMethod(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), "public static void main")
}

func GetForwardedArgs() []string {
	for i, arg := range os.Args {
		if arg == "--" {
			if i+1 < len(os.Args) {
				return os.Args[i+1:]
			}
			break
		}
	}
	return []string{}
}

func RunWorkspaceAutomatic(workspace models.WorkspaceManifest) {
    buildOrder, err := SortModules(workspace)
    if err != nil {
        log.Error("Circular dependency detected", "error", err)
        return
    }

    log.Info("Running workspace monorepo automatically...", "modules", buildOrder)
    for _, modName := range buildOrder {
        modConfig := workspace.Modules[modName]
        log.Info("Executing workspace module", "module", modName, "path", modConfig.Path)
        RunModuleProject(modConfig.Path, "", []string{}, workspace)
    }
}

func RunScript(scriptName string, scriptArgs []string, manifest *models.Manifest) error {
	if pre, ok := manifest.Scripts["pre"+scriptName]; ok {
		log.Info("Running pre-script", "script", scriptName)
		if err := executeShellCommand(pre, scriptName, nil); err != nil {
			return err
		}
	}

	mainCmd, ok := manifest.Scripts[scriptName]
	if !ok {
		log.Error("Script not found", "script", scriptName)
		return fmt.Errorf("script '%s' not found in manifest", scriptName)
	}

	log.Info("Running script", "script", scriptName, "args", scriptArgs)
	if err := executeShellCommand(mainCmd, scriptName, scriptArgs); err != nil {
		return err
	}

	if post, ok := manifest.Scripts["post"+scriptName]; ok {
		log.Info("Running post-script", "script", scriptName)
		return executeShellCommand(post, scriptName, nil)
	}

	return nil
}

func RunWorkspaceScript(workspace models.WorkspaceManifest, scriptName string, scriptArgs []string) error {
	log.Info("Running workspace-wide script across modules...", "script", scriptName)

	executedCount := 0

	for modName, modConfig := range workspace.Modules {
		manifestFiles := []string{"jar-cart.json", "jar-cart.xml"}
		var loadedManifest *models.Manifest

		for _, f := range manifestFiles {
			fullPath := filepath.Join(modConfig.Path, f)
			if adapter := GetAdapterForFile(fullPath); adapter != nil {
				if m, err := adapter.Load(fullPath); err == nil {
					loadedManifest = m
					break
				}
			}
		}

		if loadedManifest == nil || loadedManifest.Scripts == nil {
			continue
		}

		_, hasMain := loadedManifest.Scripts[scriptName]
		_, hasPre := loadedManifest.Scripts["pre"+scriptName]
		_, hasPost := loadedManifest.Scripts["post"+scriptName]

		if !hasMain && !hasPre && !hasPost {
			continue
		}

		log.Info("Executing script in module context", "module", modName, "path", modConfig.Path)

		originalDir, err := os.Getwd()
		if err != nil {
			return err
		}

		if err := os.Chdir(modConfig.Path); err != nil {
			return fmt.Errorf("failed to enter module directory %s: %w", modConfig.Path, err)
		}

		err = RunScript(scriptName, scriptArgs, loadedManifest)
		_ = os.Chdir(originalDir)

		if err != nil {
			return fmt.Errorf("script '%s' failed in module %s: %w", scriptName, modName, err)
		}

		executedCount++
	}

	if executedCount == 0 {
		log.Warn("No modules found containing the specified script", "script", scriptName)
	} else {
		log.Info("Workspace script execution completed successfully across modules", "count", executedCount)
	}

	return nil
}

func executeShellCommand(cmdStr string, eventName string, scriptArgs []string) error {
	var cmd *exec.Cmd

	_, javacPath, _, err := GetJDKPaths()
	jdkBinDir := ""
	if err == nil {
		jdkBinDir = filepath.Dir(javacPath)
	}

	exePath, err := os.Executable()
	if err != nil {
		exePath = "jar-cart"
	}

	cmdStr = strings.ReplaceAll(cmdStr, "{JAR_CART}", exePath)

	if len(scriptArgs) > 0 {
		cmdStr += " " + strings.Join(scriptArgs, " ")
	}

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	cwd, _ := os.Getwd()
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		"INIT_CWD="+cwd,
		"JAR_CART_LIFECYCLE_EVENT="+eventName,
	)

	if jdkBinDir != "" {
		cmd.Env = append(
			cmd.Env,
			"PATH="+jdkBinDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}