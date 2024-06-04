package jenkins

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/bndr/gojenkins"
)

// StringCredentials store only secret text
type StringCredentials = gojenkins.StringCredentials

// UsernameCredentials struct representing credential for storing username-password pair
type UsernameCredentials = gojenkins.UsernameCredentials

// SSHCredentials store credentials for ssh keys.
type SSHCredentials = gojenkins.SSHCredentials

// DockerServerCredentials store credentials for docker keys.
type DockerServerCredentials = gojenkins.DockerServerCredentials

// CredentialsManager is utility to control credential plugin
type CredentialsManager interface {
	Update(context.Context, string, string, interface{}) error
}

// NewCredentialsManager returns a new CredentialManager
func NewCredentialsManager(jenkins *gojenkins.Jenkins) CredentialsManager {
	return gojenkins.CredentialsManager{J: jenkins}
}

// UpdateCredential overwrites an existing credential
func UpdateCredential(ctx context.Context, credentialsManager CredentialsManager, domain string, credential interface{}) error {
	credValue := reflect.ValueOf(credential)
	if credValue.Kind() != reflect.Struct {
		return fmt.Errorf("'credential' parameter is supposed to be a Credentials struct not '%s'", credValue.Type())
	}

	idField := credValue.FieldByName("ID")
	if !idField.IsValid() || idField.Kind() != reflect.String {
		return fmt.Errorf("'credential' parameter is supposed to be a Credentials struct not '%s'", credValue.Type())
	}

	secretID := idField.String()
	if secretID == "" {
		return errors.New("secret ID should not be empty")
	}

	return credentialsManager.Update(ctx, domain, secretID, credential)
}
