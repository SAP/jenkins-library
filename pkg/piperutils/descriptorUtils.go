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

	"github.com/pkg/errors"
)

const (
	// NameRegex is used to match the pip descriptor artifact name
	NameRegex = "(?s)(.*)name=['\"](.*?)['\"](.*)"
	// VersionRegex is used to match the pip descriptor artifact version
	VersionRegex = "(?s)(.*)version=['\"](.*?)['\"](.*)"
	// MethodRegex is used to identify a method within pip descriptor to dynamically load the version from txt file
	MethodRegex = "(?s)(.*)\\(\\)"
)

// BuildDescriptor acts as a general purpose accessor to coordinates
type BuildDescriptor interface {
	GetVersion() string
	SetVersion(string)
}

// MavenDescriptor holds the unique identifier combination for Maven built Java artifacts
type MavenDescriptor struct {
	XMLName    xml.Name `xml:"project"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Packaging  string   `xml:"packaging"`
}

// GetVersion returns the Maven descriptor version property
func (desc *MavenDescriptor) GetVersion() string {
	return desc.Version
}

// SetVersion sets the Maven descriptor version property
func (desc *MavenDescriptor) SetVersion(v string) {
	desc.Version = v
}

// PipDescriptor holds the unique identifier combination for pip built Python artifacts
type PipDescriptor struct {
	GroupID    string
	ArtifactID string
	Version    string
	Packaging  string
}

// GetVersion returns the Pip descriptor version property
func (desc *PipDescriptor) GetVersion() string {
	return desc.Version
}

// SetVersion sets the Pip descriptor version property
func (desc *PipDescriptor) SetVersion(v string) {
	desc.Version = v
}

// GetMavenCoordinates reads the coordinates from the maven pom.xml descriptor file
func GetMavenCoordinates(filename string) (*MavenDescriptor, error) {
	r, _ := regexp.Compile(`\$\{.*?\}`)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read descriptor file %v", filename)
	}
	result := &MavenDescriptor{}
	err = xml.Unmarshal(data, result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse descriptor file %v", filename)
	}
	if len(result.GroupID) == 0 || r.MatchString(result.GroupID) {
		result.GroupID, err = calculateCoordinate(filename, "groupId", `(?m)^[\s*\w+\.]+`)
		if len(result.GroupID) == 0 || err != nil {
			return nil, errors.Wrap(err, "Failed to determine groupId")
		}
	}
	if len(result.ArtifactID) == 0 || r.MatchString(result.ArtifactID) {
		return nil, errors.Wrap(err, "Failed to determine artifactId")
	}
	if len(result.Version) == 0 || r.MatchString(result.Version) {
		result.Version, err = calculateCoordinate(filename, "version", `(?m)^\s*([0-9]+[\.-]*)+`)
		if len(result.Version) == 0 || err != nil {
			return nil, errors.Wrap(err, "Failed to determine version")
		}
	}
	if len(result.Packaging) == 0 || r.MatchString(result.Packaging) {
		result.Packaging, err = calculateCoordinate(filename, "packaging", `(?m)^[\s*\w+\.]+`)
		if len(result.Packaging) == 0 || err != nil {
			return nil, errors.Wrap(err, "Failed to determine packaging")
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

// GetPipCoordinates returns the pip build descriptor coordinates
func GetPipCoordinates(filename string) (*PipDescriptor, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read descriptor file %v", filename)
	}
	content := string(data)
	descriptor := &PipDescriptor{}
	if evaluateResult(content, NameRegex) {
		compile := regexp.MustCompile(NameRegex)
		values := compile.FindStringSubmatch(content)
		descriptor.ArtifactID = values[2]
	} else {
		descriptor.ArtifactID = ""
	}
	if evaluateResult(content, VersionRegex) {
		compile := regexp.MustCompile(VersionRegex)
		values := compile.FindStringSubmatch(content)
		descriptor.Version = values[2]
	} else {
		descriptor.Version = ""
	}
	if len(descriptor.Version) <= 0 || evaluateResult(descriptor.Version, MethodRegex) {
		filename = strings.Replace(filename, "setup.py", "version.txt", 1)
		descriptor.Version, err = getVersionFromFile(filename)
		if err != nil {
			return descriptor, err
		}
	}

	return descriptor, nil
}

func evaluateResult(value, regex string) bool {
	if len(value) > 0 {
		match, err := regexp.MatchString(regex, value)
		if err != nil || match {
			return true
		}
		return false
	}
	return true
}

func getVersionFromFile(file string) (string, error) {
	version, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	versionString := string(version)
	if len(versionString) >= 0 {
		return strings.TrimSpace(versionString), nil
	}
	return "", nil
}
