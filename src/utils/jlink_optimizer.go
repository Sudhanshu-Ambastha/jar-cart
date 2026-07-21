package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/charmbracelet/log"
)

type JLinkOptimizer struct {
	JarPath string
}

func (o *JLinkOptimizer) AnalyzeDependencies(javaVersion string) (string, error) {
	if _, err := exec.LookPath("jdeps"); err != nil {
		return "", fmt.Errorf("jdeps not found in PATH: %w", err)
	}

	absJarPath, err := filepath.Abs(o.JarPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve jar path: %w", err)
	}

	args := []string{
		"--print-module-deps",
		"--ignore-missing-deps",
		"-R",
		"--multi-release", javaVersion,
	}

	libDir := "lib"
	if FileExists(libDir) && hasJars(libDir) {
		absLibPath, err := filepath.Abs(libDir)
		if err == nil {
			classpath := filepath.Join(absLibPath, "*")
			args = append(args, "--class-path", classpath)
		}
	}

	args = append(args, absJarPath)

	log.Info("Analyzing with jdeps", "jar", absJarPath, "version", javaVersion)
	cmd := exec.Command("jdeps", args...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("jdeps failed: %s, error: %w", string(out), err)
	}

	output := strings.TrimSpace(string(out))
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Warning:") {
			continue
		}
		return line, nil
	}

	return "", fmt.Errorf("could not parse valid module list from jdeps output: %s", output)
}

func hasJars(dir string) bool {
	found := false
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".jar") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

func (o *JLinkOptimizer) CreateCustomRuntime(modules string, outputDir string, cfg models.OptimizationConfig) error {
	args := []string{
		"--add-modules", modules,
		"--output", outputDir,
		"--no-header-files",
		"--no-man-pages",
		"--compress=" + strconv.Itoa(cfg.Compression),
	}

	if cfg.StripDebug {
		args = append(args, "--strip-debug")
	}
	
	if cfg.StripNative {
		args = append(args, "--strip-native-debug-symbols", "exclude-files")
	}

	log.Info("Running jlink with optimization...", "compression", cfg.Compression, "strip-debug", cfg.StripDebug)
	
	cmd := exec.Command("jlink", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("jlink failed: %s, %w", string(output), err)
	}

	return nil
}