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

func RunProject(input string, appArgs []string) {
	javaVersion, javacPath, javaPath, err := GetJDKPaths()
	if err != nil {
		log.Error("JDK configuration error", "error", err)
		return
	}

	binDir, _ := filepath.Abs(filepath.Join(".jar-cart", "bin"))
	_ = os.RemoveAll(binDir)
	_ = os.MkdirAll(binDir, 0755)

	absLib, _ := filepath.Abs("lib")
	classpath := binDir + string(os.PathListSeparator) + filepath.Join(absLib, "*")

	cwd, _ := os.Getwd()

	javaFiles, err := FindJavaFiles(cwd)
	if err != nil || len(javaFiles) == 0 {
		log.Error("No Java source files detected")
		return
	}

	argfilePath := filepath.Join(".jar-cart", "sources.txt")
	argfile, err := os.Create(argfilePath)
	if err != nil {
		log.Error("Failed to create javac argfile", "error", err)
		return
	}
	defer func() {
		argfile.Close()
		_ = os.Remove(argfilePath)
	}()

	for _, file := range javaFiles {
		relPath, _ := filepath.Rel(cwd, file)
		_, _ = argfile.WriteString(relPath + "\n")
	}

	if err := argfile.Close(); err != nil {
		log.Error("Failed to finalize javac argfile", "error", err)
		return
	}

	log.Info("Compiling with JDK", "version", javaVersion)

	javacCmd := exec.Command(
		javacPath,
		"-cp",
		classpath,
		"-sourcepath",
		cwd,
		"-d",
		binDir,
		"@"+argfilePath,
	)

	javacCmd.Stdout = os.Stdout
	javacCmd.Stderr = os.Stderr

	if err := javacCmd.Run(); err != nil {
		log.Error("Compilation failed", "error", err)
		return
	}

	targetFile, err := resolveTargetFile(cwd, input, javaFiles)
	if err != nil {
		log.Error("Could not resolve target", "input", input, "error", err)
		return
	}

	mainClass := ResolveMainClass(targetFile)

	log.Info(
		"Booting execution engine",
		"jdk", javaVersion,
		"class", mainClass,
		"args", appArgs,
	)

	args := []string{
		"-cp",
		classpath,
		mainClass,
	}

	args = append(args, appArgs...)

	javaCmd := exec.Command(javaPath, args...)
	javaCmd.Stdin = os.Stdin
	javaCmd.Stdout = os.Stdout
	javaCmd.Stderr = os.Stderr

	if err := javaCmd.Run(); err != nil {
		log.Error("Execution failed", "error", err)
	}
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

	if runtime.GOOS == "windows" {
		cmdStr = strings.ReplaceAll(cmdStr, "{JAR_CART}", "\""+exePath+"\"")
	} else {
		cmdStr = strings.ReplaceAll(cmdStr, "{JAR_CART}", exePath)
	}

	if len(scriptArgs) > 0 {
		cmdStr += " " + strings.Join(scriptArgs, " ")
	}

	if runtime.GOOS == "windows" {
		cmd = exec.Command(
			"powershell",
			"-NoProfile",
			"-Command",
			"& "+cmdStr,
		)
	} else {
		cmd = exec.Command(
			"sh",
			"-c",
			cmdStr,
		)
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