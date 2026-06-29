package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
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
    if err == nil {
        lines := strings.Split(string(content), "\n")
        for _, line := range lines {
            line = strings.TrimSpace(line)
            if strings.HasPrefix(line, "package ") {
                pkg := strings.TrimPrefix(line, "package ")
                pkg = strings.TrimSuffix(pkg, ";")
                className := strings.TrimSuffix(filepath.Base(filePath), ".java")
                return pkg + "." + className
            }
        }
    }
    return strings.TrimSuffix(filepath.Base(filePath), ".java")
}

func RunProject(input string) {
	javaVersion, javacPath, javaPath, err := GetJDKPaths()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	binDir, _ := filepath.Abs(filepath.Join(".jar-cart", "bin"))
	_ = os.RemoveAll(binDir)
	_ = os.MkdirAll(binDir, 0755)

	absLib, _ := filepath.Abs("lib")
	classpath := binDir + string(os.PathListSeparator) + filepath.Join(absLib, "*")
	searchPath := input
	var mainFileToResolve string

	info, err := os.Stat(input)
	if err == nil && info.IsDir() {
		javaFiles, _ := FindJavaFiles(input)
		if len(javaFiles) == 0 {
			fmt.Printf("❌ Error: No Java source files detected in '%s/'.\n", input)
			return
		}
		mainFileToResolve = javaFiles[0] 
	} else {
		searchPath = filepath.Dir(input)
		mainFileToResolve = input
	}

	javaFiles, err := FindJavaFiles(searchPath)
	if err != nil || len(javaFiles) == 0 {
		fmt.Printf("❌ Error: No Java source files detected.\n")
		return
	}

	argfilePath := filepath.Join(".jar-cart", "sources.txt")
	argfile, _ := os.Create(argfilePath)
	for _, file := range javaFiles {
		_, _ = argfile.WriteString(filepath.Clean(file) + "\n")
	}
	argfile.Close()
	defer os.Remove(argfilePath)

	fmt.Printf("⚡ Compiling with JDK %s...\n", javaVersion)
	javacCmd := exec.Command(javacPath, "-cp", classpath, "-d", binDir, "@"+argfilePath)
	javacCmd.Stdout = os.Stdout
	javacCmd.Stderr = os.Stderr
	
	if err := javacCmd.Run(); err != nil {
		fmt.Println("❌ Compilation failed.")
		return
	}

	mainClass := ResolveMainClass(mainFileToResolve)
	
	fmt.Printf("🚀 Booting execution engine (JDK %s): %s\n\n", javaVersion, mainClass)
	
	javaCmd := exec.Command(javaPath, "-cp", classpath, mainClass)
	javaCmd.Stdin = os.Stdin
	javaCmd.Stdout = os.Stdout
	javaCmd.Stderr = os.Stderr
	
	if err := javaCmd.Run(); err != nil {
		fmt.Printf("❌ Execution failed: %v\n", err)
	}
}

func RunScript(scriptName string, manifest *models.Manifest) error {
	if pre, ok := manifest.Scripts["pre"+scriptName]; ok {
		fmt.Printf("🔍 Running pre-%s...\n", scriptName)
		if err := executeShellCommand(pre, scriptName); err != nil {
			return err
		}
	}

	if main, ok := manifest.Scripts[scriptName]; ok {
		fmt.Printf("⚡ Running %s...\n", scriptName)
		if err := executeShellCommand(main, scriptName); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("script '%s' not found in manifest", scriptName)
	}

	if post, ok := manifest.Scripts["post"+scriptName]; ok {
		fmt.Printf("🔍 Running post-%s...\n", scriptName)
		return executeShellCommand(post, scriptName)
	}
	return nil
}

func executeShellCommand(cmdStr string, eventName string) error {
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
		powershellCmd := "& " + cmdStr
		cmd = exec.Command("powershell", "-NoProfile", "-Command", powershellCmd)
	} else {
		cmdStr = strings.ReplaceAll(cmdStr, "{JAR_CART}", exePath)
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	
	cwd, _ := os.Getwd()
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, 
		"INIT_CWD="+cwd,
		"JAR_CART_LIFECYCLE_EVENT="+eventName,
	)

	if jdkBinDir != "" {
		pathSep := string(os.PathListSeparator)
		cmd.Env = append(cmd.Env, "PATH="+jdkBinDir+pathSep+os.Getenv("PATH"))
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}