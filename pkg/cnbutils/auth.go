package cnbutils

import (
	"encoding/json"
	"fmt"

	"github.com/docker/cli/cli/config/configfile"
)

func GenerateCnbAuth(config string, utils BuildUtils) (string, error) {
	var err error
	dockerConfig := &configfile.ConfigFile{}

	if config != "" {
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
		auth[registry] = fmt.Sprintf("Basic %s", value.Auth)
	}

	cnbRegistryAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(cnbRegistryAuth), nil
}
