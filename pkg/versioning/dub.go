package versioning

import (
	"io/ioutil"
	"os"
)

// Dub defines a dub artifact used for versioning
type Dub struct {
	DubJSONPath    string
	DubJSONContent map[string]interface{}
	ReadFile       func(string) ([]byte, error)
	WriteFile      func(string, []byte, os.FileMode) error
}

func (d *Dub) init() {
	if len(d.DubJSONPath) == 0 {
		d.DubJSONPath = "dub.json"
	}
	if d.ReadFile == nil {
		d.ReadFile = ioutil.ReadFile
	}

	if d.WriteFile == nil {
		d.WriteFile = ioutil.WriteFile
	}
}

// VersioningScheme returns the relevant versioning scheme
func (d *Dub) VersioningScheme() string {
	return "semver2"
}

// GetVersion returns the current version of the artifact
func (d *Dub) GetVersion() (string, error) {
	d.init()

	dubJSON := JSONfile{Path: d.DubJSONPath, ReadFile: d.ReadFile, WriteFile: d.WriteFile}

	return dubJSON.GetVersion("version")
}

// SetVersion updates the version of the artifact
func (d *Dub) SetVersion(version string) error {
	d.init()

	dubJSON := JSONfile{Path: d.DubJSONPath, ReadFile: d.ReadFile, WriteFile: d.WriteFile}

	return dubJSON.SetVersion("version", version)
}
