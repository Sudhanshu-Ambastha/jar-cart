package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func DetectManifestFile() string {
	if _, err := os.Stat("jar-cart.xml"); err == nil {
		return "jar-cart.xml"
	}
	return "jar-cart.json"
}

func GetJDKPaths() (string, string, string, error) {
	manifestFile := DetectManifestFile()
	manifest, _ := LoadManifest(manifestFile)
	
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
		return "", "", "", fmt.Errorf("JDK %s not found at %s. Run 'jar-cart sync' to provision it", javaVersion, jdkDir)
	}

	return javaVersion, javacPath, javaPath, nil
}

func GetFileHash(path string) ([32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return [32]byte{}, err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return [32]byte{}, err
	}
	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res, nil
}

func CalculateProjectHash(files []string) ([32]byte, error) {
	h := sha256.New()
	for _, f := range files {
		hash, err := GetFileHash(f)
		if err != nil {
			return [32]byte{}, err
		}
		h.Write(hash[:])
	}
	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res, nil
}