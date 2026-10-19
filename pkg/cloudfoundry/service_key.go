package cloudfoundry

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ReadServiceKey returns the JSON service key for the specified service instance.
func ReadServiceKey(runner command.ExecRunner, options ServiceKeyOptions) (serviceKey string, err error) {
	loginOptions := LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}
	if err := Login(runner, loginOptions); err != nil {
		return "", fmt.Errorf("Login to Cloud Foundry failed: %w", err)
	}
	defer func() {
		if logoutErr := Logout(runner); logoutErr != nil && err == nil {
			err = fmt.Errorf("Logout of Cloud Foundry failed: %w", logoutErr)
		}
	}()

	var output bytes.Buffer
	runner.Stdout(&output)

	log.Entry().WithField("cfServiceInstance", options.CfServiceInstance).WithField("cfServiceKey", options.CfServiceKeyName).Info("Read service key for service instance")
	if err := runner.RunExecutable("cf", "service-key", options.CfServiceInstance, options.CfServiceKeyName); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", fmt.Errorf("Reading service key failed: %w", err)
	}

	return serviceKeyJSON(output.String()), nil
}

func serviceKeyJSON(output string) string {
	lines := strings.Split(output, "\n")
	if len(lines) <= 2 {
		return ""
	}
	return strings.Join(lines[2:], "")
}

// ServiceKeyOptions identifies a Cloud Foundry service key and the target where it resides.
type ServiceKeyOptions struct {
	CfAPIEndpoint     string
	CfOrg             string
	CfSpace           string
	CfServiceInstance string
	CfServiceKeyName  string
	Username          string
	Password          string
}
