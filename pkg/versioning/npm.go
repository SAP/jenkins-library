package versioning

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

// Npm ...
type Npm struct {
	PackageJSONPath    string
	PackageJSONContent map[string]interface{}
	ReadFile           func(string) ([]byte, error)
	WriteFile          func(string, []byte, os.FileMode) error
}

// InitBuildDescriptor ...
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

// VersioningScheme ...
func (n *Npm) VersioningScheme() string {
	return "semver2"
}

// GetVersion ...
func (n *Npm) GetVersion() (string, error) {
	n.init()

	content, err := n.ReadFile(n.PackageJSONPath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read file '%v'", n.PackageJSONPath)
	}

	err = json.Unmarshal(content, &n.PackageJSONContent)
	if err != nil {
		return "", errors.Wrap(err, "failed to read package.json content")
	}

	return fmt.Sprint(n.PackageJSONContent["version"]), nil
}

// SetVersion ...
func (n *Npm) SetVersion(version string) error {
	n.init()

	if n.PackageJSONContent == nil {
		_, err := n.GetVersion()
		if err != nil {
			return err
		}
	}
	n.PackageJSONContent["version"] = version

	content, err := json.MarshalIndent(n.PackageJSONContent, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to create json content for package.json")
	}
	err = n.WriteFile(n.PackageJSONPath, content, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to write file '%v'", n.PackageJSONPath)
	}

	return nil
}
