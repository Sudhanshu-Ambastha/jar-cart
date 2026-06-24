package models

import (
	"encoding/xml"
)

type Dependency struct {
	Group   string `json:"group" xml:"group"`
	Library string `json:"library" xml:"library"`
	Version string `json:"version" xml:"version"`
}

type Manifest struct {
	XMLName      xml.Name     `json:"-" xml:"Manifest"`
	Project      string       `json:"project" xml:"project"`
	Strategy     string       `json:"strategy" xml:"strategy"`
	JavaVersion  string       `json:"java_version" xml:"java_version"`
	Dependencies []Dependency `json:"dependencies" xml:"dependencies>dependency"`
}