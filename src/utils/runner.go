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

func ResolveMainClass(input string) string {
	if input == "src" || input == "." {
		return "src.App"
	}
	
	name := strings.TrimSuffix(input, ".java")
	name = strings.ReplaceAll(name, string(os.PathSeparator), ".")
	name = strings.ReplaceAll(name, "/", ".")
	
	if !strings.HasPrefix(name, "src.") {
		name = "src." + name
	}
	
	return name
}

func RunProject(input string) {
	mainClass := ResolveMainClass(input)
	
	binDir := filepath.Join(".jar-cart", "bin")
	absLib, err := filepath.Abs("lib")
	if err != nil {
		return
	}
	
	_ = os.MkdirAll(binDir, 0755)
	classpath := binDir + string(os.PathListSeparator) + filepath.Join(absLib, "*")

	javaFiles, err := FindJavaFiles("src")
	if err != nil || len(javaFiles) == 0 {
		fmt.Println("❌ Error: No Java source files detected in 'src/'.")
		return
	}

	argfilePath := filepath.Join(".jar-cart", "sources.txt")
	argfile, _ := os.Create(argfilePath)
	for _, file := range javaFiles {
		_, _ = argfile.WriteString(filepath.Clean(file) + "\n")
	}
	argfile.Close()
	defer os.Remove(argfilePath)

	fmt.Println("⚡ Compiling source tree architecture...")
	
	javacCmd := exec.Command("javac", "-cp", classpath, "-d", binDir, "@"+argfilePath)
	javacCmd.Stdout = os.Stdout
	javacCmd.Stderr = os.Stderr
	
	if err := javacCmd.Run(); err != nil {
		fmt.Println("❌ Compilation failed.")
		return
	}

	fmt.Printf("🚀 Booting execution engine: %s\n\n", mainClass)

	javaCmd := exec.Command("java", "-cp", classpath, mainClass)
	javaCmd.Stdin = os.Stdin
	javaCmd.Stdout = os.Stdout
	javaCmd.Stderr = os.Stderr
	env := append(os.Environ(), "JAVA_TOOL_OPTIONS=--enable-native-access=ALL-UNNAMED")
	javaCmd.Env = env
	
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
	cmd.Env = append(os.Environ(), 
		"INIT_CWD="+cwd,
		"JAR_CART_LIFECYCLE_EVENT="+eventName,
	)
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}