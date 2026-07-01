package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func GetJDKPaths() (string, string, string, error) {
	manifest, _ := LoadManifest("jar-cart.json")
	javaVersion := "25" 
	if manifest != nil && manifest.JavaVersion != "" {
		javaVersion = manifest.JavaVersion
	}

	home, _ := os.UserHomeDir()
	jdkDir := filepath.Join(home, ".jar-cart", "jdks", javaVersion)
	
	javacPath := filepath.Join(jdkDir, "bin", "javac")
	javaPath := filepath.Join(jdkDir, "bin", "java")
	
	if runtime.GOOS == "windows" {
		javacPath += ".exe"
		javaPath += ".exe"
	}

	if _, err := os.Stat(javacPath); err != nil {
		return "", "", "", fmt.Errorf("JDK %s not found at %s. Run 'jar-cart sync'", javaVersion, jdkDir)
	}

	return javaVersion, javacPath, javaPath, nil
}