package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

type DecompilerTool struct {
	Name       string
	GroupID    string
	ArtifactID string
}

var TrustedDecompilers = map[string]DecompilerTool{
	"vineflower": {
		Name:       "Vineflower",
		GroupID:    "org.vineflower",
		ArtifactID: "vineflower",
	},
	"cfr": {
		Name:       "CFR",
		GroupID:    "org.benf",
		ArtifactID: "cfr",
	},
	"procyon": {
		Name:       "Procyon",
		GroupID:    "org.bitbucket.mstrobel",
		ArtifactID: "procyon-decompiler",
	},
}

func Decompile(jarPath string, engine string) error {
	tool, ok := TrustedDecompilers[engine]
	if !ok {
		return fmt.Errorf("unsupported decompiler engine: %s", engine)
	}

	log.Info("Preparing decompiler", "engine", tool.Name)
	version := GetLatestVersionFromMaven(tool.GroupID, tool.ArtifactID)
	if version == "" {
		return fmt.Errorf("could not resolve latest version for %s", tool.Name)
	}

	home, _ := os.UserHomeDir()
	toolDir := filepath.Join(home, ".jar-cart", "tools", engine)
	toolJar := filepath.Join(toolDir, fmt.Sprintf("%s-%s.jar", tool.ArtifactID, version))

	if !FileExists(toolJar) {
		log.Info("Downloading decompiler", "engine", tool.Name, "version", version)
		os.MkdirAll(toolDir, 0755)
		
		groupPath := strings.ReplaceAll(tool.GroupID, ".", "/")
		url := fmt.Sprintf("https://repo1.maven.org/maven2/%s/%s/%s/%s-%s.jar",
			groupPath, tool.ArtifactID, version, tool.ArtifactID, version)
		
		if err := downloadFile(url, toolJar); err != nil {
			return fmt.Errorf("failed to download decompiler: %w", err)
		}
	}

	outputDir := "decompiled_" + engine
	os.MkdirAll(outputDir, 0755)
	
	log.Info("Running decompiler", "engine", tool.Name)
	return runCommand("java", "-jar", toolJar, jarPath, outputDir)
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}