package adapters

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sudhanshu-Ambastha/jar-cart/src/models"
)

type GradleAdapter struct {
	BaseAdapter
}

func (a *GradleAdapter) Load(path string) (*models.Manifest, error) {
	return &models.Manifest{
		Project:         "gradle-project", 
		ResolutionDepth: "full", 
		Scripts:         make(map[string]string),
	}, nil
}

func (a *GradleAdapter) Save(path string, m *models.Manifest) error {
	return nil 
}

func (a *GradleAdapter) AddDependency(path string, dep models.Dependency) error {
	content, err := os.ReadFile(path)
	if err != nil { return err }
	
	depLine := fmt.Sprintf("    implementation '%s:%s:%s'\n", dep.Group, dep.Library, dep.Version)
	
	newContent := strings.Replace(string(content), "dependencies {", "dependencies {\n"+depLine, 1)
	return os.WriteFile(path, []byte(newContent), 0644)
}

func (a *GradleAdapter) RemoveDependency(path string, dep models.Dependency, libDir string) error {
	a.DeleteJar(libDir, dep)
	return nil
}

func (a *GradleAdapter) Sync(path string, deps []models.Dependency) error {
	content, err := os.ReadFile(path)
	if err != nil { return err }

	var sb strings.Builder
	sb.WriteString("dependencies {\n")
	for _, d := range deps {
		sb.WriteString(fmt.Sprintf("    implementation '%s:%s:%s'\n", d.Group, d.Library, d.Version))
	}
	sb.WriteString("}")

	newContent := ""
	if strings.Contains(string(content), "dependencies {") {
		newContent = string(content) 
	} else {
		newContent = string(content) + "\n\n" + sb.String()
	}
	return os.WriteFile(path, []byte(newContent), 0644)
}