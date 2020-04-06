package versioning

import (
	"io/ioutil"
	"os"
)

// Npm defines an npm artifact used for versioning
type Npm struct {
	PackageJSONPath    string
	PackageJSONContent map[string]interface{}
	ReadFile           func(string) ([]byte, error)
	WriteFile          func(string, []byte, os.FileMode) error
}

func (n *Npm) init() {
	if len(n.PackageJSONPath) == 0 {
		n.PackageJSONPath = "package.json"
	}
	if n.ReadFile == nil {
		n.ReadFile = ioutil.ReadFile
	}

	if n.WriteFile == nil {
		n.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (n *Npm) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact
func (n *Npm) GetVersion() (string, error) {
	n.init()

	packageJSON := JSONfile{Path: n.PackageJSONPath, ReadFile: n.ReadFile, WriteFile: n.WriteFile}

	return packageJSON.GetVersion("version")
}

// SetVersion updates the version of the artifact
func (n *Npm) SetVersion(version string) error {
	n.init()

	packageJSON := JSONfile{Path: n.PackageJSONPath, ReadFile: n.ReadFile, WriteFile: n.WriteFile}

	return packageJSON.SetVersion("version", version)
}
