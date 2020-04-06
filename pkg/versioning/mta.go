package versioning

import (
	"io/ioutil"
	"os"
)

// Mta defines a mta artifact used for versioning
type Mta struct {
	MtaYAMLPath    string
	MtaYAMLContent map[string]interface{}
	ReadFile       func(string) ([]byte, error)
	WriteFile      func(string, []byte, os.FileMode) error
}

func (m *Mta) init() {
	if len(m.MtaYAMLPath) == 0 {
		m.MtaYAMLPath = "mta.yaml"
	}
	if m.ReadFile == nil {
		m.ReadFile = ioutil.ReadFile
	}

	if m.WriteFile == nil {
		m.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (m *Mta) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact
func (m *Mta) GetVersion() (string, error) {
	m.init()

	mtaYAML := YAMLfile{Path: m.MtaYAMLPath, ReadFile: m.ReadFile, WriteFile: m.WriteFile}

	return mtaYAML.GetVersion("version")
}

// SetVersion updates the version of the artifact
func (m *Mta) SetVersion(version string) error {
	m.init()

	mtaYAML := YAMLfile{Path: m.MtaYAMLPath, ReadFile: m.ReadFile, WriteFile: m.WriteFile}

	return mtaYAML.SetVersion("version", version)
}
