package adapters

import (
	"encoding/json"
	"os"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type JsonAdapter struct { BaseAdapter }

func (a *JsonAdapter) Load(path string) (*models.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &models.Manifest{Dependencies: []models.Dependency{}, Scripts: make(map[string]string)}, nil
	}
	if len(data) == 0 {
		return &models.Manifest{Dependencies: []models.Dependency{}, Scripts: make(map[string]string)}, nil
	}

	var m models.Manifest
	err = json.Unmarshal(data, &m)

	if m.Scripts == nil {
		m.Scripts = make(map[string]string)
	}
	
	return &m, err
}

func (a *JsonAdapter) Save(path string, m *models.Manifest) error {
	data, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (a *JsonAdapter) AddDependency(path string, dep models.Dependency) error {
	m, _ := a.Load(path)
	m.Dependencies = append(m.Dependencies, dep)
	return a.Save(path, m)
}

func (a *JsonAdapter) RemoveDependency(path string, dep models.Dependency, libDir string) error {
	m, _ := a.Load(path)
	m.Dependencies = a.FilterDependencies(m.Dependencies, dep)
	a.DeleteJar(libDir, dep)
	return a.Save(path, m)
}

func (a *JsonAdapter) Sync(path string, dependencies []models.Dependency) error {
	m, _ := a.Load(path)
	m.Dependencies = dependencies
	return a.Save(path, m)
}