package protecode

import (
	"encoding/base64"
	"fmt"
	"testing"
)

func TestDockerConfigAuthWithHappyCase(t *testing.T) {
	expected := "user1:pass1"
	secret := base64.StdEncoding.EncodeToString([]byte(expected))
	testAuth := DockerConfigAuth{secret}
	cred, err := testAuth.encodedAuth()
	if err != nil {
		t.Errorf("expected no error with %s, but got %s", secret, err)
	}

	if cred != expected {
		t.Errorf("expected Auth to be %s, but got %s", expected, cred)
	}
}

func TestDockerConfigPrintsRedactedString(t *testing.T) {
	expected := "user1:pass1"
	secret := base64.StdEncoding.EncodeToString([]byte(expected))
	testAuth := DockerConfigAuth{secret}
	expectedAuth := "Auth: ***"
	redactedAuth := fmt.Sprintf("%s", testAuth)
	if redactedAuth != expectedAuth {
		t.Errorf("expected auth be redacted as %s, but got %s", expectedAuth, redactedAuth)
	}
}

func TestDockerConfigAuthWithEmptyString(t *testing.T) {
	secret := ""
	testAuth := DockerConfigAuth{secret}
	if _, err := testAuth.encodedAuth(); err != nil {
		t.Errorf("Expected no error, but got %s", err)
	}
}

func TestNewDockerConfigFromJSONWithHappyCase(t *testing.T) {
	testHost := "example.com"
	expectedAuth := "dXNlcjE6cGFzczEK"
	testDockerConfigJSON := `{"auths": {"example.com": {"auth": "dXNlcjE6cGFzczEK"}}}`

	config, err := NewDockerConfigFromJSON(testDockerConfigJSON)
	if err != nil {
		t.Errorf("Expected no errors with happy case, but got %s", err)
	}

	if len(config.Auths) != 1 {
		t.Errorf("Expected config.Auths to have exactly 1 item, but got %v", len(config.Auths))
	}

	hostConfig, ok := config.Auths[testHost]
	if !ok {
		t.Errorf("Expected config.Auths to include %s, but got nothing", hostConfig)
	}

	if hostConfig.Auth != expectedAuth {
		t.Errorf("Expected Host Auth to be %s, but got %s", expectedAuth, hostConfig.Auth)
	}
}

func TestNewDockerConfigFromJSONRaiseErrorWhenEmptyString(t *testing.T) {
	if _, err := NewDockerConfigFromJSON(""); err == nil {
		t.Errorf("Expected error with DockerConfigJSON is empty string")
	}
}
