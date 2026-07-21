package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
	"github.com/charmbracelet/log"
)

func OrchestrateBuild(changedFilePath string) {
	workspace := LoadWorkspaceManifest()
	buildOrder, err := SortModules(workspace)
	if err != nil {
		log.Error("Circular dependency detected", "error", err)
		return
	}

	for _, modName := range buildOrder {
		modConfig := workspace.Modules[modName]

		if IsModuleDirty(modConfig.Path) {
			log.Info("Module dirty, orchestrating...", "module", modName)
			RunProject(modConfig.Path, []string{})
			UpdateModuleHash(modConfig.Path)
		}
	}
}

func IsModuleDirty(modulePath string) bool {
	files := GetSourceFiles(modulePath)
	currentHash, err := CalculateProjectHash(files)
	if err != nil {
		return true
	}
	lastHash := LoadSavedHash(modulePath)
	return string(currentHash[:]) != string(lastHash[:])
}

func LoadWorkspaceManifest() models.WorkspaceManifest {
	data, err := os.ReadFile("jar-cart.workspace.json")
	if err != nil {
		log.Fatal("Could not find jar-cart.workspace.json in root")
	}
	var wm models.WorkspaceManifest
	json.Unmarshal(data, &wm)
	return wm
}

func GetSourceFiles(modulePath string) []string {
	var files []string
	filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".java") {
			files = append(files, path)
		}
		return nil
	})
	return files
}

func UpdateModuleHash(modulePath string) {
	files := GetSourceFiles(modulePath)
	hash, err := CalculateProjectHash(files)
	if err != nil {
		return
	}
	dir := filepath.Join(modulePath, ".jar-cart")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "last_build.hash"), hash[:], 0644)
}

func LoadSavedHash(modulePath string) [32]byte {
	var hash [32]byte
	data, err := os.ReadFile(filepath.Join(modulePath, ".jar-cart", "last_build.hash"))
	if err != nil {
		return [32]byte{}
	}
	copy(hash[:], data)
	return hash
}