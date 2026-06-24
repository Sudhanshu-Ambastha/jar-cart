package adapters

import (
	"os"
	"path/filepath"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type BaseAdapter struct{}

func (b *BaseAdapter) FilterDependencies(current []models.Dependency, target models.Dependency) []models.Dependency {
	var updated []models.Dependency
	for _, d := range current {
		if d.Group != target.Group || d.Library != target.Library || d.Version != target.Version {
			updated = append(updated, d)
		}
	}
	return updated
}

func (b *BaseAdapter) DeleteJar(libDir string, dep models.Dependency) {
	jarName := dep.Library + "-" + dep.Version + ".jar"
	os.Remove(filepath.Join(libDir, jarName))
}