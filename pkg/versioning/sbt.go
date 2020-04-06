package versioning

import (
	"io/ioutil"
	"os"
)

// Sbt defines an npm artifact used for versioning
type Sbt struct {
	DescriptorPath    string
	DescriptorContent map[string]interface{}
	ReadFile          func(string) ([]byte, error)
	WriteFile         func(string, []byte, os.FileMode) error
}

func (s *Sbt) init() {
	if len(s.DescriptorPath) == 0 {
		s.DescriptorPath = "sbtDescriptor.json"
	}
	if s.ReadFile == nil {
		s.ReadFile = ioutil.ReadFile
	}

	if s.WriteFile == nil {
		s.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (s *Sbt) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact
func (s *Sbt) GetVersion() (string, error) {
	s.init()

	packageJSON := JSONfile{Path: s.DescriptorPath, ReadFile: s.ReadFile, WriteFile: s.WriteFile}

	return packageJSON.GetVersion("version")
}

// SetVersion updates the version of the artifact
func (s *Sbt) SetVersion(version string) error {
	s.init()

	packageJSON := JSONfile{Path: s.DescriptorPath, ReadFile: s.ReadFile, WriteFile: s.WriteFile}

	return packageJSON.SetVersion("version", version)
}
