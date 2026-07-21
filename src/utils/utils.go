package utils

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	depsdev "deps.dev/api/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/charmbracelet/log"
)

type MavenResolver struct{}

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

func GetExcludedModules() []string {
    content, err := os.ReadFile(".jarcartignore")
    if err != nil {
        return []string{} 
    }
    
    var modules []string
    lines := strings.Split(string(content), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && !strings.HasPrefix(line, "#") {
            modules = append(modules, line)
        }
    }
    return modules
}

func TrackDuration(start time.Time, name string, logger *log.Logger) {
    elapsed := time.Since(start)
    logger.Info("Command finished", "cmd", name, "duration", elapsed.Round(time.Millisecond).String())
}

func SortModules(workspace models.WorkspaceManifest) ([]string, error) {
	order := []string{}
	visited := make(map[string]bool)
	temp := make(map[string]bool)

	var visit func(string) error
	visit = func(name string) error {
		if temp[name] {
			return fmt.Errorf("circular dependency detected at: %s", name)
		}
		if !visited[name] {
			temp[name] = true
			if mod, ok := workspace.Modules[name]; ok {
				for _, dep := range mod.DependsOn {
					if _, ok := workspace.Modules[dep]; ok {
						if err := visit(dep); err != nil {
							return err
						}
					}
				}
			}
			temp[name] = false
			visited[name] = true
			order = append(order, name)
		}
		return nil
	}

	for name := range workspace.Modules {
		if err := visit(name); err != nil {
			return nil, err
		}
	}
	return order, nil
}

func extractGroup(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func extractArtifact(name string) string {
	parts := strings.Split(name, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func ResolveUsingGoogle(deps []models.Dependency) ([]models.Dependency, error) {
    conn, err := grpc.Dial("api.deps.dev:443", grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
    if err != nil {
        return nil, err
    }
    defer conn.Close()
    client := depsdev.NewInsightsClient(conn)

    var resolved []models.Dependency
    for _, d := range deps {
        resp, err := client.GetDependencies(context.Background(), &depsdev.GetDependenciesRequest{
            VersionKey: &depsdev.VersionKey{
                System:  depsdev.System_MAVEN,
                Name:    d.Group + ":" + d.Library,
                Version: d.Version,
            },
        })
        if err != nil {
            return nil, err
        }
        
        resolved = append(resolved, d)
        
        for _, node := range resp.Nodes {
            resolved = append(resolved, models.Dependency{
                Group:   extractGroup(node.VersionKey.Name),
                Library: extractArtifact(node.VersionKey.Name),
                Version: node.VersionKey.Version,
            })
        }
    }
    return resolved, nil
}

func BuildWorkspaceClasspath(workspace models.WorkspaceManifest, currentModulePath string) string {
    buildOrder, err := SortModules(workspace)
    if err != nil {
        return ""
    }

    var classpathEntries []string
    for _, modName := range buildOrder {
        modConfig := workspace.Modules[modName]
        if modConfig.Path == currentModulePath {
            break
        }
        outDir, _ := filepath.Abs(filepath.Join(modConfig.Path, ".jar-cart", "bin"))
        classpathEntries = append(classpathEntries, outDir)
    }

    return strings.Join(classpathEntries, string(os.PathListSeparator))
}

func getDirectorySize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}