package piperutils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	piperCmd "github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

// MavenGAV holds the unique identifier combination for maven artifacts
type MavenGAV struct {
	XMLName    xml.Name `xml:"project"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
}

// GetMavenGAV reads the coordinates from the maven pom.xml descriptor file
func GetMavenGAV(filename string) (*MavenGAV, error) {
	r, _ := regexp.Compile(`\$\{.*?\}`)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to read descriptor file %v: %v", filename, err)
	}
	result := &MavenGAV{}
	err = xml.Unmarshal(data, result)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse descriptor file %v: %v", filename, err)
	}
	if len(result.GroupID) == 0 || r.MatchString(result.GroupID) {
		result.GroupID, err = calculateCoordinate(filename, "groupId", `(?m)^[\s*\w+\.]+`)
		if len(result.GroupID) == 0 || err != nil {
			return nil, err
		}
	}
	if len(result.ArtifactID) == 0 || r.MatchString(result.ArtifactID) {
		result.ArtifactID, err = calculateCoordinate(filename, "artifactId", `(?m)^[\s*\w+\.]+`)
		if len(result.ArtifactID) == 0 || err != nil {
			return nil, err
		}
	}
	if len(result.Version) == 0 || r.MatchString(result.Version) {
		result.Version, err = calculateCoordinate(filename, "version", `(?m)^\s*([0-9]+[\.-]*)+`)
		if len(result.Version) == 0 || err != nil {
			return nil, fmt.Errorf("Failed to determine version: %v", err)
		}
	}
	return result, nil
}

func calculateCoordinate(filename, coordinate, filterRegex string) (string, error) {
	output := &bytes.Buffer{}
	cmd := piperCmd.Command{}
	cmd.Stdout(output)
	err := cmd.RunExecutable("mvn", "-f", filename, fmt.Sprintf(`-Dexpression=project.%v`, coordinate), "help:evaluate")
	stdout := output.String()
	if err != nil {
		return "", fmt.Errorf("Failed to calculate coordinate version on descriptor %v: %v (error %v)", filename, stdout, err)
	}
	log.Entry().WithField("package", "github.com/SAP/jenkins-library/pkg/piperutils").Debugf("Maven output was: %v", stdout)
	return filter(stdout, filterRegex), nil
}

func filter(text, filterRegex string) string {
	r, _ := regexp.Compile(filterRegex)
	return strings.TrimSpace(r.FindString(text))
}
