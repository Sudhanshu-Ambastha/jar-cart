package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/utils/adapters"
	"github.com/manifoldco/promptui"
)

func LoadManifest(filePath string) (*models.Manifest, error) {
	adapter := GetAdapterForFile(filePath)
	if adapter == nil {
		return nil, fmt.Errorf("unsupported manifest format: %s", filepath.Ext(filePath))
	}
	return adapter.Load(filePath)
}

func ParseManifest(filePath string) ([]models.Dependency, error) {
	manifest, err := LoadManifest(filePath)
	if err != nil {
		return nil, err
	}
	return manifest.Dependencies, nil
}

func GenerateLockFile(projectDir string, deps []models.Dependency) error {
	fmt.Println("🔒 Generating/Updating lockfile...")
	lockEntries, err := ResolveParallelDependencies(projectDir, deps)
	if err != nil {
		return err
	}

	lock := LockFile{
		Version:      1,
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Dependencies: lockEntries,
	}
	return WriteLockFile(projectDir, &lock)
}

func updateOrAddDependency(manifest *models.Manifest, newDep models.Dependency) bool {
	for i, d := range manifest.Dependencies {
		if d.Group == newDep.Group && d.Library == newDep.Library {
			if manifest.Dependencies[i].Version != newDep.Version {
				fmt.Printf("🔄 Updating version: %s -> %s\n", d.Version, newDep.Version)
				manifest.Dependencies[i].Version = newDep.Version
				return true
			}
			fmt.Println("ℹ️ Dependency already exists. Checking synchronization...")
			return false
		}
	}
	manifest.Dependencies = append(manifest.Dependencies, newDep)
	return true
}

func refreshLockFile() error {
	fmt.Println("🔄 Regenerating lockfile...")
	newLock := &LockFile{
		Version:      1,
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Dependencies: make(map[string]LockEntry),
	}

	files, err := os.ReadDir("lib")
	if err != nil {
		return err
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".jar") {
			fullPath := filepath.Join("lib", f.Name())
			hash, _ := CalculateSHA256(fullPath)
			info, _ := f.Info()
			newLock.Dependencies[f.Name()] = LockEntry{
				Path:   fullPath,
				Size:   info.Size(),
				SHA256: hash,
			}
		}
	}
	return WriteLockFile(".", newLock)
}

func AddDependency(manifestPath, rawCoordinate string, isDirect bool, libDir string) error {
    group, lib, version, err := resolveCoordinate(rawCoordinate)
    if err != nil {
        return err
    }

    newDep := models.Dependency{Group: group, Library: lib, Version: version}
    adapter := GetAdapterForFile(manifestPath)
    manifest, err := adapter.Load(manifestPath)
    if err != nil {
        return fmt.Errorf("failed to load manifest: %w", err)
    }

    modified := updateOrAddDependency(manifest, newDep)

    if modified {
        if err := adapter.Save(manifestPath, manifest); err != nil {
            return fmt.Errorf("failed to save manifest: %w", err)
        }
        fmt.Println("💾 Manifest updated successfully.")
    } else {
        fmt.Println("ℹ️ Manifest already contains this dependency.")
    }

    fmt.Printf("🔒 Synchronizing: %s:%s\n", group, lib)
    if _, err := ResolveParallelDependencies(".", []models.Dependency{newDep}); err != nil {
        return err
    }

    if err := refreshLockFile(); err != nil {
        fmt.Printf("⚠️ Warning: Failed to update lockfile: %v\n", err)
    } else {
        fmt.Println("✅ Lockfile generated successfully.")
    }

    return nil
}

func RemoveDependency(manifestPath, rawCoordinate string, libDir string) error {
    adapter := GetAdapterForFile(manifestPath)
    if adapter == nil {
        return fmt.Errorf("no adapter found for %s", manifestPath)
    }

    manifest, err := adapter.Load(manifestPath)
    if err != nil {
        return err
    }

    var targetDep *models.Dependency
    for _, d := range manifest.Dependencies {
        if d.Library == rawCoordinate || fmt.Sprintf("%s:%s", d.Group, d.Library) == rawCoordinate {
            targetDep = &d
            break
        }
    }

    if targetDep == nil {
        return fmt.Errorf("dependency '%s' not found in manifest", rawCoordinate)
    }

    err = adapter.RemoveDependency(manifestPath, *targetDep, libDir)
    if err != nil {
        return err
    }

    manifest, _ = adapter.Load(manifestPath)
    return GenerateLockFile(".", manifest.Dependencies)
}

func RunSync(projectDir string) error {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %v", err)
	}

	manifestFiles := []string{"jar-cart.json", "jar-cart.xml", "build.gradle", "pom.xml"}
	var manifestPath string

	for _, f := range manifestFiles {
		fullPath := filepath.Join(absDir, f)
		if _, err := os.Stat(fullPath); err == nil {
			manifestPath = f
			break
		}
	}

	if manifestPath == "" {
		return fmt.Errorf("no manifest file found in %s", absDir)
	}

	fmt.Printf("📦 Synchronizing via: %s\n", manifestPath)
	deps, err := ParseManifest(filepath.Join(absDir, manifestPath))
	if err != nil {
		return fmt.Errorf("parse manifest error: %v", err)
	}

	lockEntries, err := ResolveParallelDependencies(absDir, deps)
	if err != nil {
		return fmt.Errorf("resolve error: %v", err)
	}

	err = CleanupLibDir(absDir, lockEntries)
	if err != nil {
		return fmt.Errorf("cleanup error: %v", err)
	}

	fmt.Println("✨ Dependencies synced and linked perfectly!")
	return nil
}
func syncToProjectFiles(deps []models.Dependency) {
	targets := map[string]adapters.ManifestAdapter{
		"build.gradle": &adapters.GradleAdapter{},
	}

	for fileName, adapter := range targets {
		if _, err := os.Stat(fileName); err == nil {
			fmt.Printf("🔄 Adaptive Sync: Updating native file -> %s\n", fileName)
			_ = adapter.Sync(fileName, deps)
		}
	}
}

func resolveCoordinate(raw string) (string, string, string, error) {
	parts := strings.Split(raw, ":")
	if len(parts) < 2 {
		fmt.Printf("🔍 Ambiguous coordinate '%s'. Searching for matches...\n", raw)
		suggestions := GetSearchSuggestions(raw)
		if len(suggestions) == 0 {
			return "", "", "", fmt.Errorf("no matches found for '%s'", raw)
		}
		prompt := promptui.Select{
			Label: "Select the correct artifact",
			Items: suggestions,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "❯ {{ .G }}:{{ .A }} (latest: {{ .LatestVersion }})",
				Inactive: "  {{ .G }}:{{ .A }}",
				Selected: "✔ Selected {{ .G }}:{{ .A }}",
			},
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return "", "", "", err
		}
		sel := suggestions[idx]
		return sel.G, sel.A, sel.LatestVersion, nil
	}
	group, lib := parts[0], parts[1]
	version := ""
	if len(parts) == 3 {
		version = parts[2]
	}
	if version == "" || version == "+" {
		version = GetLatestVersionFromMaven(group, lib)
	}
	return group, lib, version, nil
}

func ConvertManifest(sourcePath, targetExt string) error {
	srcAdapter := GetAdapterForFile(sourcePath)
	if srcAdapter == nil {
		return fmt.Errorf("unsupported source format: %s", sourcePath)
	}

	manifest, err := srcAdapter.Load(sourcePath)
	if err != nil {
		return err
	}

	basePath := strings.TrimSuffix(sourcePath, filepath.Ext(sourcePath))
	targetPath := basePath + "." + targetExt

	var targetAdapter adapters.ManifestAdapter
	switch strings.ToLower(targetExt) {
	case "json":
		targetAdapter = &adapters.JsonAdapter{}
	case "xml":
		targetAdapter = &adapters.XmlAdapter{}
	default:
		return fmt.Errorf("unsupported target format: %s", targetExt)
	}

	err = targetAdapter.Save(targetPath, manifest)
	if err != nil {
		return err
	}

	fmt.Printf("🧹 Removing old manifest: %s\n", sourcePath)
	return os.Remove(sourcePath)
}

func GetAdapterForFile(path string) adapters.ManifestAdapter {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return &adapters.JsonAdapter{}
	case ".xml":
		return &adapters.XmlAdapter{}
	case ".gradle":
		return &adapters.GradleAdapter{}
	default:
		return nil
	}
}