package npm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultConfigFilename = ".piperNpmrc" // default by npm
)

var (
	propertiesLoadFile  = ioutil.ReadFile
	propertiesWriteFile = ioutil.WriteFile
)

func NewNPMRC(path string) NPMRC {
	if !strings.HasSuffix(path, defaultConfigFilename) {
		path = filepath.Join(path, defaultConfigFilename)
	}

	return NPMRC{filepath: path}
}

type NPMRC struct {
	filepath string
	content  string
}

func (rc *NPMRC) Write() error {
	if err := propertiesWriteFile(rc.filepath, []byte(rc.content), 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", rc.filepath)
	}
	return nil
}

func (rc *NPMRC) Load() error {
	bytes, err := propertiesLoadFile(rc.filepath)
	if err != nil {
		return err
	}
	rc.content = string(bytes)
	return nil
}

func (rc *NPMRC) Set(key, value string) {
	r := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*%s\s*=.*$`, key))

	keyValue := fmt.Sprintf("%s=%s", key, value)

	if r.MatchString(rc.content) {
		rc.content = r.ReplaceAllString(rc.content, keyValue)
	} else {
		rc.content += keyValue + "\n"
	}
}
