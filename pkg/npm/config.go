package npm

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/magiconair/properties"
	"github.com/pkg/errors"
)

const (
	configFilename  = ".npmrc"
	authKeyTemplate = "//%s:_authToken"
)

func encode(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
}

//     //registry.npmjs.org/:_authToken=${NPM_TOKEN}
func (rc NPMRC) SetAuthToken(registryUrl, token string) {
	registryUrl = strings.TrimPrefix(registryUrl, "https://")
	registryUrl = strings.TrimPrefix(registryUrl, "http://")
	rc.Set(fmt.Sprintf(authKeyTemplate, registryUrl), token)
}
func (rc NPMRC) SetAuth(registryUrl, username, password string) {
	rc.SetAuthToken(registryUrl, encode(username, password))
}

func (rc NPMRC) Set(key, value string) {
	rc.values.Set(key, value)
}

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

func (rc NPMRC) Write() error {
	file, err := os.OpenFile(rc.filepath /*os.O_APPEND|*/, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s", rc.filepath)
	}
	defer file.Close()
	_, err = file.WriteString(rc.values.String())
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", rc.filepath)
	}
	return nil
}

func (rc NPMRC) Load() error {
	log.Entry().Debugf("loading existing file %s", rc.filepath)
	values, err := properties.LoadFile(rc.filepath, properties.UTF8)
	log.Entry().Debug("content: %s", values.String())
	if err != nil {
		return err
	}
	rc.values = values
	return err
}

func (rc NPMRC) Print() string {
	return rc.values.String()
}
