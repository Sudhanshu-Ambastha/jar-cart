package adapters

import (
	"encoding/xml"
	"os"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type XmlAdapter struct {
	BaseAdapter
}

func (a *XmlAdapter) Load(path string) (*models.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &models.Manifest{Dependencies: []models.Dependency{}, Scripts: make(map[string]string)}, nil
	}

	var m models.Manifest
	err = xml.Unmarshal(data, &m)
	
	m.Scripts = make(map[string]string)
	for _, s := range m.XMLScripts {
		m.Scripts[s.Name] = s.Command
	}
	return &m, err
}

func (a *XmlAdapter) Save(path string, m *models.Manifest) error {
	if !strings.Contains(path, "pom.xml") {
		m.XMLScripts = []models.Script{}
		for k, v := range m.Scripts {
			m.XMLScripts = append(m.XMLScripts, models.Script{Name: k, Command: v})
		}
	} else {
		m.XMLScripts = nil
	}
	
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