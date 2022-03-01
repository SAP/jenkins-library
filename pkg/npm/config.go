package npm

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
	"github.com/pkg/errors"
)

const (
	defaultConfigFilename = ".piperNpmrc" // default by npm
)

var (
	propertiesLoadFile  = properties.LoadFile
	propertiesWriteFile = ioutil.WriteFile
)

func NewNPMRC(path string) NPMRC {
	if !strings.HasSuffix(path, defaultConfigFilename) {
		path = filepath.Join(path, defaultConfigFilename)
	}

	return NPMRC{filepath: path, values: properties.NewProperties()}
}

type NPMRC struct {
	filepath string
	values   *properties.Properties
}

func (rc *NPMRC) Write() error {
	if err := propertiesWriteFile(rc.filepath, []byte(rc.values.String()), 0644); err != nil {
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
