package adapters

import "github.com/Sudhanshu-Ambastha/jar-cart/src/models"

type ManifestAdapter interface {
	Load(path string) (*models.Manifest, error)
	Save(path string, m *models.Manifest) error
	AddDependency(path string, dep models.Dependency) error
	RemoveDependency(path string, dep models.Dependency, libDir string) error
	Sync(path string, dependencies []models.Dependency) error
}