package maven

import (
	"encoding/xml"
	"fmt"
)

// Project describes the Maven object model
type Project struct {
	XMLName      xml.Name     `xml:"project"`
	Parent       Parent       `xml:"parent"`
	GroupId      string       `xml:"groupId"`
	ArtifactId   string       `xml:"artifactId"`
	Version      string       `xml:"version"`
	Packaging    string       `xml:"packaging"`
	Name         string       `xml:"name"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
	Modules      []string     `xml:"modules>module"`
}

// Parent describes the coordinates a module's parent POM
type Parent struct {
	XMLName    xml.Name `xml:"parent"`
	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
	Version    string   `xml:"version"`
}

// Dependency describes a dependency of the module
type Dependency struct {
	XMLName    xml.Name    `xml:"dependency"`
	GroupId    string      `xml:"groupId"`
	ArtifactId string      `xml:"artifactId"`
	Version    string      `xml:"version"`
	Classifier string      `xml:"classifier"`
	Type       string      `xml:"type"`
	Scope      string      `xml:"scope"`
	Exclusions []Exclusion `xml:"exclusions>exclusion"`
}

// Exclusion describes an exclusion within a dependency
type Exclusion struct {
	XMLName    xml.Name `xml:"exclusion"`
	GroupId    string   `xml:"groupId"`
	ArtifactId string   `xml:"artifactId"`
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
