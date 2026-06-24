package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

func DownloadJars(dependencies []models.Dependency, libDir string) error {
	if len(dependencies) == 0 {
		return nil
	}

	err := os.MkdirAll(libDir, 0755)
	if err != nil {
		return fmt.Errorf("unable to construct local library path target workspace: %w", err)
	}

	baseMavenURL := "https://repo1.maven.org/maven2"

	for _, d := range dependencies {
		jarFileName := fmt.Sprintf("%s-%s.jar", d.Library, d.Version)
		localFilePath := filepath.Join(libDir, jarFileName)

		if _, err := os.Stat(localFilePath); err == nil {
			fmt.Printf("ℹ️ Dependency cache hit: %s is already present inside /%s\n", jarFileName, libDir)
			continue
		}

		groupPath := strings.ReplaceAll(d.Group, ".", "/")
		targetURL := fmt.Sprintf("%s/%s/%s/%s/%s", baseMavenURL, groupPath, d.Library, d.Version, jarFileName)
		fmt.Printf("📥 Fetching: %s...\n", jarFileName)
		
		resp, err := http.Get(targetURL)
		if err != nil {
			return fmt.Errorf("failed fetching connection pipeline context for %s: %w", jarFileName, err)
		}
		
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("server returned status %d for %s (URL tried: %s)", resp.StatusCode, jarFileName, targetURL)
		}

		out, err := os.Create(localFilePath)
		if err != nil {
			resp.Body.Close()
			return fmt.Errorf("failed creating disk sector anchor for %s: %w", jarFileName, err)
		}

		_, err = io.Copy(out, resp.Body)
		out.Close()
		resp.Body.Close()
		
		if err != nil {
			return fmt.Errorf("stream breakdown writing payload context data to disk for %s: %w", jarFileName, err)
		}

		fmt.Printf("✨ Download complete: %s successfully cached in /%s\n", jarFileName, libDir)
	}

	return nil
}