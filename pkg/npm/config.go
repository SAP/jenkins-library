package npm

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
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
	if !strings.HasPrefix(path, configFilename) {
		path = filepath.Join(path, configFilename)
	}
	return NPMRC{path: path, values: properties.NewProperties()}
}

type NPMRC struct {
	path   string
	values *properties.Properties
}

func (rc NPMRC) Write() error {
	file, err := os.OpenFile(rc.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	_, err = file.WriteString(rc.values.String())
	if err != nil {
		return err
	}
	return nil
}

func (rc NPMRC) Load() {
	rc.values = properties.MustLoadFile(rc.path, properties.UTF8)
}
