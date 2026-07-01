package models

import (
	"encoding/xml"
)

type Dependency struct {
	Group   string `json:"group" xml:"group"`
	Library string `json:"library" xml:"library"`
	Version string `json:"version" xml:"version"`
}

type Script struct {
	Name    string `xml:"name,attr"`
	Command string `xml:",chardata"`
}

type Manifest struct {
	XMLName      xml.Name          `json:"-" xml:"Manifest"`
	Project      string            `json:"project" xml:"project"`
	JavaVersion  string            `json:"java_version" xml:"java_version"`
	ResolutionDepth string            `json:"resolution_depth"`
	Scripts      map[string]string `json:"scripts,omitempty" xml:"-"`
	Dependencies []Dependency      `json:"dependencies" xml:"dependencies>dependency"`
	XMLScripts   []Script          `json:"-" xml:"scripts>script"`
}

type Pom struct {
	XMLName      xml.Name `xml:"project"`
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
		Scope      string `xml:"scope"`
	} `xml:"dependencies>dependency"`
}