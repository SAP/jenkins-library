package protecode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// DockerConfigAuth holds Auth details for each Host
type DockerConfigAuth struct {
	Auth string
}

func (ac *DockerConfigAuth) encodedAuth() (string, error) {
	token, err := base64.StdEncoding.DecodeString(ac.Auth)
	if err != nil {
		return "", errors.New("Failed to decode base64 secret")
	}

	return string(token), nil
}

// To avoid accidental leaks in logs
func (ac DockerConfigAuth) String() string {
	return fmt.Sprintf("Auth: ***")
}

// DockerConfig presentation
type DockerConfig struct {
	Auths map[string]DockerConfigAuth
}

// NewDockerConfigFromJSON builds DockerConfig from JSON string
func NewDockerConfigFromJSON(jsonText string) (DockerConfig, error) {
	if len(jsonText) == 0 {
		return DockerConfig{}, errors.New("DockerConfigJSON is empty")
	}

	var dc DockerConfig
	if err := json.Unmarshal([]byte(jsonText), &dc); err != nil {
		return DockerConfig{}, errors.New("Failed to parse DockerConfig from given JSON string")
	}

	return dc, nil
}

func (dc *DockerConfig) getHostAuth(host string) (DockerConfigAuth, bool) {
	if hostAuth, ok := dc.Auths[host]; ok == true {
		return hostAuth, true
	}

	return DockerConfigAuth{}, false
}
