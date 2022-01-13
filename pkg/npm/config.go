package npm

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
	"github.com/pkg/errors"
)

const (
	configFilename = ".piperNpmrc"
)

var (
	propertiesLoadFile = properties.LoadFile
)

func NewNPMRC(path string) NPMRC {
	if !strings.HasSuffix(path, configFilename) {
		path = filepath.Join(path, configFilename)
	}
	return NPMRC{filepath: path, values: properties.NewProperties()}
}

type NPMRC struct {
	filepath string
	values   *properties.Properties
}

func (rc *NPMRC) Write() error {
	if err := ioutil.WriteFile(rc.filepath, []byte(rc.values.String()), 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", rc.filepath)
	}
	return nil
}

func (rc *NPMRC) Load() error {
	values, err := propertiesLoadFile(rc.filepath, properties.UTF8)
	if err != nil {
		return err
	}
	rc.values = values
	return nil
}

func (rc *NPMRC) Set(key, value string) {
	rc.values.Set(key, value)
}
