package versioning

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Docker defines an artifact based on a Dockerfile
type Docker struct {
	artifact         Artifact
	content          []byte
	utils            Utils
	options          *Options
	path             string
	versionSource    string
	versioningScheme string
	readFile         func(string) ([]byte, error)
	writeFile        func(string, []byte, os.FileMode) error
}

func (d *Docker) init() {
	if d.readFile == nil {
		d.readFile = ioutil.ReadFile
	}

	if d.writeFile == nil {
		d.writeFile = ioutil.WriteFile
	}
}

func (d *Docker) initDockerfile() {
	if len(d.path) == 0 {
		d.path = "Dockerfile"
	}
}

// VersioningScheme returns the relevant versioning scheme
func (d *Docker) VersioningScheme() string {
	if len(d.versioningScheme) == 0 {
		return "docker"
	}
	return d.versioningScheme
}

// GetVersion returns the current version of the artifact
func (d *Docker) GetVersion() (string, error) {
	d.init()
	var err error

	switch d.versionSource {
	case "FROM":
		var err error
		d.initDockerfile()
		d.content, err = d.readFile(d.path)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read file '%v'", d.path)
		}
		version := d.versionFromBaseImageTag()
		if len(version) == 0 {
			return "", fmt.Errorf("no version information available in FROM statement")
		}
		return version, nil
	case "":
		if len(d.path) == 0 {
			d.path = "VERSION"
		}
		d.versionSource = "custom"
		fallthrough
	case "custom", "dub", "golang", "maven", "mta", "npm", "pip", "sbt":
		if d.options == nil {
			d.options = &Options{}
		}
		d.artifact, err = GetArtifact(d.versionSource, d.path, d.options, d.utils)
		if err != nil {
			return "", err
		}
		return d.artifact.GetVersion()
	default:
		d.initDockerfile()
		d.content, err = d.readFile(d.path)
		if err != nil {
			return "", errors.Wrapf(err, "failed to read file '%v'", d.path)
		}
		version := d.versionFromEnv(d.versionSource)
		if len(version) == 0 {
			return "", fmt.Errorf("no version information available in ENV '%v'", d.versionSource)
		}
		return version, nil
	}
}

// SetVersion updates the version of the artifact
func (d *Docker) SetVersion(version string) error {
	d.init()

	dir := ""

	if d.artifact != nil {
		err := d.artifact.SetVersion(version)
		if err != nil {
			return err
		}
		dir = filepath.Dir(d.path)
	}

	err := d.writeFile(filepath.Join(dir, "VERSION"), []byte(version), 0700)
	if err != nil {
		return errors.Wrap(err, "failed to write file 'VERSION'")
	}

	return nil
}

func (d *Docker) versionFromEnv(env string) string {
	lines := strings.Split(string(d.content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ENV") && strings.Fields(line)[1] == env {
			return strings.Fields(line)[2]
		}
	}
	return ""
}

func (d *Docker) versionFromBaseImageTag() string {
	lines := strings.Split(string(d.content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "FROM") {
			imageParts := strings.Split(line, ":")
			partsCount := len(imageParts)
			if partsCount == 1 {
				return ""
			}
			version := imageParts[partsCount-1]
			return strings.TrimSpace(version)
		}
	}
	return ""
}

// GetCoordinates returns the coordinates
func (d *Docker) GetCoordinates() (Coordinates, error) {
	return nil, nil
}
