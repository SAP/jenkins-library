package cmd

import (
	"github.com/SAP/jenkins-library/pkg/cloudfoundry"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func cloudFoundryCreateSpace(config cloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData) {

	c := command.Command{}
	cf := cloudfoundry.CFUtils{Exec: &command.Command{}}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	err := runCloudFoundryCreateSpace(&config, telemetryData, cf, &c)

	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}

}

func runCloudFoundryCreateSpace(config *cloudFoundryCreateSpaceOptions, telemetryData *telemetry.CustomData, cf cloudfoundry.CFUtils, s command.ShellRunner) (err error) {
	return cloudfoundry.CreateSpace((*cloudfoundry.CloudFoundryCreateSpaceOptions)(config), telemetryData, cf, s)
}
