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

func RunProject(target string) {
	binDir := filepath.Join(".jar-cart", "bin")
	absLib, _ := filepath.Abs("lib")
	libPath := filepath.Join(absLib, "*")
	_ = os.MkdirAll(binDir, 0755)

	classpathSep := string(os.PathListSeparator)

	javaFiles, err := FindJavaFiles("src")
	if err != nil || len(javaFiles) == 0 {
		fmt.Println("❌ Error: No Java source files detected in 'src/'.")
		return
	}

	argfilePath := filepath.Join(".jar-cart", "sources.txt")
	argfile, _ := os.Create(argfilePath)
	for _, file := range javaFiles {
		_, _ = argfile.WriteString(file + "\n")
	}
	argfile.Close()
	defer os.Remove(argfilePath)

	fmt.Println("⚡ Compiling source tree architecture...")
	javacArgs := []string{
		"-cp", fmt.Sprintf("%s%s%s", libPath, classpathSep, binDir),
		"-d", binDir,
		"@" + argfilePath,
	}

	javacCmd := exec.Command("javac", javacArgs...)
	javacCmd.Stdout = os.Stdout
	javacCmd.Stderr = os.Stderr
	if err := javacCmd.Run(); err != nil {
		fmt.Println("❌ Compilation lifecycle interrupted.")
		return
	}

	mainClass := "src.App"

	fmt.Printf("🚀 Booting execution engine target: %s\n\n", mainClass)
	javaArgs := []string{
		"-cp", fmt.Sprintf("%s%s%s", binDir, classpathSep, libPath),
		mainClass,
	}

	javaCmd := exec.Command("java", javaArgs...)
	javaCmd.Stdin = os.Stdin
	javaCmd.Stdout = os.Stdout
	javaCmd.Stderr = os.Stderr
	_ = javaCmd.Run()
}