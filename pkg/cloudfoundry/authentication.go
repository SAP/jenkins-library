package cloudfoundry

import (
	"errors"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

var errMissingAuth = errors.New("Parameters missing. Please provide the Cloud Foundry Endpoint, Org, Space, Username and Password")

// Login logs the current user in to Cloud Foundry with username/password credentials.
func Login(runner command.ExecRunner, options LoginOptions) error {
	if err := validateLoginOptions(options); err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}

	log.Entry().Info("Logging in to Cloud Foundry")
	log.Entry().WithField("cfAPI:", options.CfAPIEndpoint).WithField("cfOrg", options.CfOrg).WithField("space", options.CfSpace).Info("Logging into Cloud Foundry..")

	if err := loginWithCredentials(runner, options); err != nil {
		return fmt.Errorf("Failed to login to Cloud Foundry: %w", err)
	}

	log.Entry().Info("Logged in successfully to Cloud Foundry..")
	return nil
}

func validateLoginOptions(options LoginOptions) error {
	if options.CfAPIEndpoint == "" || options.CfOrg == "" || options.CfSpace == "" || options.Username == "" || options.Password == "" {
		return errMissingAuth
	}
	return nil
}

func loginWithCredentials(runner command.ExecRunner, options LoginOptions) error {
	args := append([]string{
		"login",
		"-a", options.CfAPIEndpoint,
		"-o", options.CfOrg,
		"-s", options.CfSpace,
		"-u", options.Username,
		"-p", options.Password,
	}, options.CfLoginOpts...)

	return runner.RunExecutable("cf", args...)
}

// Logout logs the current user out of Cloud Foundry.
func Logout(runner command.ExecRunner) error {
	log.Entry().Info("Logging out of Cloud Foundry")

	if err := runner.RunExecutable("cf", "logout"); err != nil {
		return fmt.Errorf("Failed to Logout of Cloud Foundry: %w", err)
	}

	log.Entry().Info("Logged out successfully")
	return nil
}

// LoginOptions contains the Cloud Foundry target and authentication details.
type LoginOptions struct {
	CfAPIEndpoint string
	CfOrg         string
	CfSpace       string
	Username      string
	Password      string
	CfLoginOpts   []string
}
