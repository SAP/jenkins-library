package versioning

import (
	"io/ioutil"
	"os"
)

// Pip defines a dub artifact used for versioning
type Pip struct {
	VersionPath string
	ReadFile    func(string) ([]byte, error)
	WriteFile   func(string, []byte, os.FileMode) error
}

func (p *Pip) init() {
	if len(p.VersionPath) == 0 {
		p.VersionPath = "version.txt"
	}
	if p.ReadFile == nil {
		p.ReadFile = ioutil.ReadFile
	}

	if p.WriteFile == nil {
		p.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (p *Pip) VersioningScheme() string {
	return "pep440"
}

// GetVersion returns the current version of the artifact
func (p *Pip) GetVersion() (string, error) {
	p.init()

	versionTxt := Versionfile{Path: p.VersionPath, ReadFile: p.ReadFile, WriteFile: p.WriteFile}

	return versionTxt.GetVersion()
}

// SetVersion updates the version of the artifact
func (p *Pip) SetVersion(version string) error {
	p.init()

	versionTxt := Versionfile{Path: p.VersionPath, ReadFile: p.ReadFile, WriteFile: p.WriteFile}

	return versionTxt.SetVersion(version)
}
