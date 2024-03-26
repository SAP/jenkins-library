package cnbutils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/docker/cli/cli/config/configfile"
)

func GenerateCnbAuth(config string, utils BuildUtils) (string, error) {
	var err error
	dockerConfig := &configfile.ConfigFile{}

	if config != "" {
		log.Entry().Debugf("using docker config file %q", config)
		dockerConfigJSON, err := utils.FileRead(config)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(dockerConfigJSON, dockerConfig)
		if err != nil {
			return "", err
		}
	}

	auth := map[string]string{}
	for registry, value := range dockerConfig.AuthConfigs {
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
