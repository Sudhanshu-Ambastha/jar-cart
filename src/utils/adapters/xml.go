package adapters

import (
	"encoding/xml"
	"os"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type XmlAdapter struct {
	BaseAdapter
}

func (a *XmlAdapter) Load(path string) (*models.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &models.Manifest{Dependencies: []models.Dependency{}}, nil
	}
	if len(data) == 0 {
		return &models.Manifest{Dependencies: []models.Dependency{}}, nil
	}

	var m models.Manifest
	err = xml.Unmarshal(data, &m)
	return &m, err
}

func (a *XmlAdapter) Save(path string, m *models.Manifest) error {
	data, err := xml.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}
	finalData := append([]byte(xml.Header), data...)
	return os.WriteFile(path, finalData, 0644)
}

func (a *XmlAdapter) AddDependency(path string, dep models.Dependency) error {
	m, err := a.Load(path)
	if err != nil {
		return err
	}
	m.Dependencies = append(m.Dependencies, dep)
	return a.Save(path, m)
}

func (a *XmlAdapter) RemoveDependency(path string, dep models.Dependency, libDir string) error {
	m, err := a.Load(path)
	if err != nil {
		return err
	}
	m.Dependencies = a.FilterDependencies(m.Dependencies, dep)
	a.DeleteJar(libDir, dep)
	return a.Save(path, m)
}

func (a *XmlAdapter) Sync(path string, dependencies []models.Dependency) error {
	m, _ := a.Load(path)
	m.Dependencies = dependencies
	return a.Save(path, m)
}