package cloudfoundry

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ReadServiceKey reads a cloud foundry service key based on provided service instance and service key name parameters
func (cf *CFUtils) ReadServiceKey(options ServiceKeyOptions) (string, error) {

	_c := cf.Exec

	if _c == nil {
		_c = &command.Command{}
	}
	cfconfig := LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}
	err := cf.Login(cfconfig)

	if err != nil {
		// error while trying to run cf login
		return "", fmt.Errorf("Login to Cloud Foundry failed: %w", err)
	}
	var serviceKeyBytes bytes.Buffer
	_c.Stdout(&serviceKeyBytes)

	// we are logged in --> read service key
	log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKeyName).Info("Read service key for service instance")
	cfReadServiceKeyScript := []string{"service-key", options.CfServiceInstance, options.CfServiceKeyName}
	err = _c.RunExecutable("cf", cfReadServiceKeyScript...)

	if err != nil {
		// error while reading service key
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Reading service key failed: %w", err)
	}

	// parse and return service key
	var serviceKeyJSON string
	if len(serviceKeyBytes.String()) > 0 {
		var lines []string = strings.Split(serviceKeyBytes.String(), "\n")
		serviceKeyJSON = strings.Join(lines[2:], "")
	}

	err = cf.Logout()
	if err != nil {
		return serviceKeyJSON, fmt.Errorf("Logout of Cloud Foundry failed: %w", err)
	}

	return serviceKeyJSON, err
}

// ServiceKeyOptions for reading CF Service Key
type ServiceKeyOptions struct {
	CfAPIEndpoint     string
	CfOrg             string
	CfSpace           string
	CfServiceInstance string
	CfServiceKeyName  string
	Username          string
	Password          string
}
