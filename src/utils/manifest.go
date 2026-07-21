package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/Sudhanshu-Ambastha/jar-cart/src/utils/adapters"
	"github.com/charmbracelet/log"
	"github.com/manifoldco/promptui"
)

func LoadManifest(filePath string) (*models.Manifest, error) {
	adapter := GetAdapterForFile(filePath)
	if adapter == nil {
		return nil, fmt.Errorf("unsupported manifest format: %s", filepath.Ext(filePath))
	}
	return adapter.Load(filePath)
}

func IsFullResolution(manifest *models.Manifest) bool {
	mode := strings.ToLower(strings.TrimSpace(manifest.ResolutionDepth))
	if mode == "" {
		mode = "full"
	}
	return mode == "full"
}

func ParseManifest(filePath string) ([]models.Dependency, error) {
	manifest, err := LoadManifest(filePath)
	if err != nil {
		return nil, err
	}
	return manifest.Dependencies, nil
}

func GenerateLockFile(projectDir string, manifest *models.Manifest) error {
	log.Info("Generating/Updating lockfile...")
	log.Info("Lockfile resolution mode: full") 
	lockEntries, err := ResolveParallelDependencies(
		projectDir,
		manifest.Dependencies,
		true,
	)
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
            if d.Version != newDep.Version {
                log.Info("Updating dependency", "artifact", d.Group+":"+d.Library, "old", d.Version, "new", newDep.Version)
                manifest.Dependencies[i].Version = newDep.Version
                return true
            }
            log.Info("Dependency already exists with same version. No changes needed.")
            return false
        }
    }
    manifest.Dependencies = append(manifest.Dependencies, newDep)
    log.Info("Added new dependency", "artifact", newDep.Group+":"+newDep.Library, "version", newDep.Version)
    return true
}

func refreshLockFile() error {
	log.Info("Regenerating lockfile...")
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
		log.Info("Manifest updated successfully")
	}

	log.Info("Synchronizing dependency", "dep", group+":"+lib)
	if _, err := ResolveParallelDependencies(libDir, []models.Dependency{newDep}, true); err != nil {
		return err
	}

	if err := refreshLockFile(); err != nil {
		log.Warn("Failed to update lockfile", "error", err)
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

    manifest, err = adapter.Load(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to reload manifest: %w", err)
	}

	return GenerateLockFile(".", manifest)
}

func RunSync(projectDir string) error {
	absDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %v", err)
	}

	manifestFiles := []string{"jar-cart.json", "jar-cart.xml"}
	var manifestPath string
	for _, f := range manifestFiles {
		fullPath := filepath.Join(absDir, f)
		if _, err := os.Stat(fullPath); err == nil {
			manifestPath = f
			break
		}
	}

	if manifestPath == "" {
		return fmt.Errorf("not found the jar-cart.json/jar-cart.xml")
	}

	log.Info("Synchronizing via", "manifest", manifestPath)
	manifest, err := LoadManifest(filepath.Join(absDir, manifestPath))
	if err != nil {
		return fmt.Errorf("load manifest error: %v", err)
	}

	libDir := filepath.Join(absDir, "lib")
	log.Info("Targeting lib directory", "path", libDir)
	
	if err := os.MkdirAll(libDir, 0755); err != nil {
		return fmt.Errorf("failed to create lib directory: %v", err)
	}

	lockEntries, err := ResolveParallelDependencies(libDir, manifest.Dependencies, true)
	if err != nil {
		return fmt.Errorf("resolve error: %v", err)
	}

	err = CleanupLibDir(absDir, lockEntries)
	if err != nil {
		return fmt.Errorf("cleanup error: %v", err)
	}

	log.Info("Dependencies synced and linked perfectly!")
	return nil
}

func syncToProjectFiles(deps []models.Dependency) {
	targets := map[string]adapters.ManifestAdapter{
		"build.gradle": &adapters.GradleAdapter{},
	}

	for fileName, adapter := range targets {
		if _, err := os.Stat(fileName); err == nil {
			log.Info("Adaptive Sync: Updating native file", "file", fileName)
			_ = adapter.Sync(fileName, deps)
		}
	}
}

func resolveCoordinate(raw string) (string, string, string, error) {
	parts := strings.Split(raw, ":")
	if len(parts) < 2 {
		log.Warn("Ambiguous coordinate, searching for matches", "input", raw)
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

	log.Info("Removing old manifest", "file", sourcePath)
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