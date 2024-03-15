package cnbutils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/docker/cli/cli/config/configfile"
)

type Credentials struct {
	utils        BuildUtils
	dockerConfig *configfile.ConfigFile
}

func NewCredentials(utils BuildUtils) Credentials {
	return Credentials{
		utils:        utils,
		dockerConfig: nil,
	}
}

func (c *Credentials) GenerateCredentials(config string) (string, error) {
	var err error
	c.dockerConfig, err = c.parse(config)
	if err != nil {
		return "", err
	}

	return c.generate()
}

func (c *Credentials) Validate(target string) bool {
	if c.dockerConfig == nil {
		return false
	}
	_, ok := c.dockerConfig.AuthConfigs[target]
	if !strings.HasPrefix(target, "localhost") && !ok {
		return false
	}
	return true
}

func (c *Credentials) parse(config string) (*configfile.ConfigFile, error) {
	dockerConfig := &configfile.ConfigFile{}

	if config != "" {
		log.Entry().Debugf("using docker config file %q", config)
		dockerConfigJSON, err := c.utils.FileRead(config)
		if err != nil {
			return &configfile.ConfigFile{}, err
		}

		err = json.Unmarshal(dockerConfigJSON, dockerConfig)
		if err != nil {
			return &configfile.ConfigFile{}, err
		}
	}
	return dockerConfig, nil
}

func (c *Credentials) generate() (string, error) {
	auth := map[string]string{}
	for registry, value := range c.dockerConfig.AuthConfigs {
		if value.Auth == "" && value.Username == "" && value.Password == "" {
			log.Entry().Warnf("docker config.json contains empty credentials for registry %q. Either 'auth' or 'username' and 'password' have to be provided.", registry)
			continue
		}

		if value.Auth == "" {
			value.Auth = encodeAuth(value.Username, value.Password)
		}

		log.Entry().Debugf("Adding credentials for: registry %q", registry)
		auth[registry] = fmt.Sprintf("Basic %s", value.Auth)
	}

	if len(auth) == 0 {
		log.Entry().Warn("docker config file is empty!")
	}

	cnbRegistryAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(cnbRegistryAuth), nil
}

func encodeAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
