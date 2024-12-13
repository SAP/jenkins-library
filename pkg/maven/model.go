package maven

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"slices"
)

// Project describes the Maven object model.
type Project struct {
	XMLName      xml.Name     `xml:"project"`
	Parent       Parent       `xml:"parent"`
	GroupID      string       `xml:"groupId"`
	ArtifactID   string       `xml:"artifactId"`
	Version      string       `xml:"version"`
	Packaging    string       `xml:"packaging"`
	Name         string       `xml:"name"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
	Modules      []string     `xml:"modules>module"`
}

// Parent describes the coordinates a module's parent POM.
type Parent struct {
	XMLName    xml.Name `xml:"parent"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
}

// Dependency describes a dependency of the module.
type Dependency struct {
	XMLName    xml.Name    `xml:"dependency"`
	GroupID    string      `xml:"groupId"`
	ArtifactID string      `xml:"artifactId"`
	Version    string      `xml:"version"`
	Classifier string      `xml:"classifier"`
	Type       string      `xml:"type"`
	Scope      string      `xml:"scope"`
	Exclusions []Exclusion `xml:"exclusions>exclusion"`
}

// Exclusion describes an exclusion within a dependency.
type Exclusion struct {
	XMLName    xml.Name `xml:"exclusion"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
}

// ParsePOM parses the provided XML raw data into a Project.
func ParsePOM(xmlData []byte) (*Project, error) {
	project := Project{}
	err := xml.Unmarshal(xmlData, &project)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POM data: %w", err)
	}
	return &project, nil
}

// ModuleInfo describes a location and Project of a maven module.
type ModuleInfo struct {
	PomXMLPath string
	Project    *Project
}

type visitUtils interface {
	FileExists(path string) (bool, error)
	FileRead(path string) ([]byte, error)
}

// VisitAllMavenModules ...
func VisitAllMavenModules(path string, utils visitUtils, excludes []string, callback func(info ModuleInfo) error) error {
	pomXMLPath := filepath.Join(path, "pom.xml")
	if slices.Contains(excludes, pomXMLPath) {
		return nil
	}

	exists, _ := utils.FileExists(pomXMLPath)
	if !exists {
		return nil
	}

	pomXMLContents, err := utils.FileRead(pomXMLPath)
	if err != nil {
		return fmt.Errorf("failed to read file contents of '%s': %w", pomXMLPath, err)
	}

	project, err := ParsePOM(pomXMLContents)
	if err != nil {
		return fmt.Errorf("failed to parse file contents of '%s': %w", pomXMLPath, err)
	}

	err = callback(ModuleInfo{PomXMLPath: pomXMLPath, Project: project})
	if err != nil {
		return err
	}

	if len(project.Modules) == 0 {
		return nil
	}

	for _, module := range project.Modules {
		subPomPath := filepath.Join(path, module)
		err = VisitAllMavenModules(subPomPath, utils, excludes, callback)
		if err != nil {
			return err
		}
	}
	return nil
}
