package models

import (
	"encoding/xml"
	"time"
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

type OptimizationConfig struct {
    Compression int  `json:"compression" xml:"compression"`
    StripDebug  bool `json:"strip_debug" xml:"strip_debug"`
    StripNative bool `json:"strip_native" xml:"strip_native"`
}

type Manifest struct {
	XMLName      xml.Name          `json:"-" xml:"Manifest"`
	Project      string            `json:"project" xml:"project"`
	JavaVersion  string            `json:"java_version" xml:"java_version"`
	ResolutionDepth string            `json:"resolution_depth"`
	Scripts      map[string]string `json:"scripts,omitempty" xml:"-"`
	Dependencies []Dependency      `json:"dependencies" xml:"dependencies>dependency"`
	XMLScripts   []Script          `json:"-" xml:"scripts>script"`
	Optimize     OptimizationConfig `json:"optimize" xml:"optimize"`
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

type ModuleConfig struct {
	Path      string   `json:"path"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

type WorkspaceManifest struct {
	Modules map[string]ModuleConfig `json:"modules"`
}

type BuildStatus struct {
	LastBuildHash string    `json:"last_build_hash"`
	LastRunTime   time.Time `json:"last_run_time"`
}

type WorkspaceState struct {
	ModuleStatuses map[string]BuildStatus `json:"module_statuses"`
}