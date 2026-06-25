package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	// If the user inputs "src" or "src/", they are likely trying to run the project.
	// We default to "src.App" as the standard entry point.
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