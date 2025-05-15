package cnbutils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	"github.com/docker/docker/registry"
)

type DockerKeychain struct {
	dockerConfig *configfile.ConfigFile
}

func (dk *DockerKeychain) ToCNBString() (string, error) {
	if dk.dockerConfig == nil || len(dk.dockerConfig.GetAuthConfigs()) == 0 {
		return "{}", nil
	}

	cnbAuth := map[string]string{}
	for reg, authConf := range dk.dockerConfig.GetAuthConfigs() {
		registryHostname := registry.ConvertToHostname(reg)
		log.Entry().Debugf("adding credentials for registry %q", registryHostname)
		if authConf.RegistryToken != "" {
			cnbAuth[registryHostname] = fmt.Sprintf("Bearer %s", authConf.RegistryToken)

			continue
		}

		if authConf.Auth != "" {
			cnbAuth[registryHostname] = fmt.Sprintf("Basic %s", authConf.Auth)

			continue
		}

		if authConf.Username == "" && authConf.Password == "" {
			log.Entry().Warnf("docker config.json contains empty credentials for registry %q. Either 'auth' or 'username' and 'password' have to be provided.", registryHostname)

			continue
		}

		cnbAuth[registryHostname] = fmt.Sprintf("Basic %s", encodeAuth(authConf.Username, authConf.Password))
	}

	cnbAuthBytes, err := json.Marshal(&cnbAuth)
	return string(cnbAuthBytes), err
}

func (dk *DockerKeychain) AuthExistsForImage(image string) bool {
	var empty types.AuthConfig
	conf, err := dk.dockerConfig.GetAuthConfig(registry.ConvertToHostname(image))
	if err != nil {
		log.Entry().Errorf("failed to get auth config for the image %q, error: %s", image, err.Error())
	}

	return conf != empty
}

func ParseDockerConfig(config string, utils BuildUtils) (*DockerKeychain, error) {
	keychain := &DockerKeychain{
		dockerConfig: &configfile.ConfigFile{},
	}
	if config != "" {
		log.Entry().Debugf("using docker config file %q", config)
		dockerConfigJSON, err := utils.FileRead(config)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(dockerConfigJSON, keychain.dockerConfig)
		if err != nil {
			return nil, err
		}
	}

	return keychain, nil
}

func encodeAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
