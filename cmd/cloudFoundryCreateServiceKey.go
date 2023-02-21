package cmd

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

const cfCliSynchronousRequestFlag = "--wait"

func cloudFoundryCreateServiceKey(options cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	cfUtils := cloudfoundry.CFUtils{
		Exec: &c,
	}

	err := runCloudFoundryCreateServiceKey(&options, telemetryData, &c, &cfUtils)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Error occurred during step.")
	}
}

func runCloudFoundryCreateServiceKey(options *cloudFoundryCreateServiceKeyOptions, telemetryData *telemetry.CustomData, c command.ExecRunner, cfUtils cloudfoundry.AuthenticationUtils) (returnedError error) {

	// Login via cf cli
	config := cloudfoundry.LoginOptions{
		CfAPIEndpoint: options.CfAPIEndpoint,
		CfOrg:         options.CfOrg,
		CfSpace:       options.CfSpace,
		Username:      options.Username,
		Password:      options.Password,
	}
	loginErr := cfUtils.Login(config)
	if loginErr != nil {
		return fmt.Errorf("Error while logging in occurred: %w", loginErr)
	}
	defer func() {
		logoutErr := cfUtils.Logout()
		if logoutErr != nil && returnedError == nil {
			returnedError = fmt.Errorf("Error while logging out occurred: %w", logoutErr)
		}
	}()
	log.Entry().Info("Creating Service Key")

	var cfCreateServiceKeyScript []string
	// the --wait option was added for cf cli v8 in order to ensure a synchronous creation of the servie key that was default in v7 or earlier
	if options.CfServiceKeyConfig == "" {
		cfCreateServiceKeyScript = []string{"create-service-key", options.CfServiceInstance, options.CfServiceKeyName}
	} else {
		cfCreateServiceKeyScript = []string{"create-service-key", options.CfServiceInstance, options.CfServiceKeyName, "-c", options.CfServiceKeyConfig}
	}

	// If a synchronous execution is requested, the "--wait" flag needs to be added
	if !options.CfAsync {
		cfCreateServiceKeyScript = append(cfCreateServiceKeyScript, cfCliSynchronousRequestFlag)
	}

	err := c.RunExecutable("cf", cfCreateServiceKeyScript...)
	if err != nil {
		return fmt.Errorf("Failed to Create Service Key: %w", err)
	}

	return returnedError
}
