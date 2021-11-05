package cmd

import (
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitPublishTargetVector(config abapAddonAssemblyKitPublishTargetVectorOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}
	maxRuntimeInMinutes := time.Duration(5 * time.Minute)
	pollIntervalsInSeconds := time.Duration(30 * time.Second)

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitPublishTargetVector(&config, telemetryData, &client, maxRuntimeInMinutes, pollIntervalsInSeconds)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitPublishTargetVector(config *abapAddonAssemblyKitPublishTargetVectorOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) error {

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client); err != nil {
		return err
	}
	conn.MaxRuntimeInMinutes = maxRuntimeInMinutes
	conn.PollIntervalsInSeconds = pollIntervalsInSeconds

	addonDescriptor := new(abaputils.AddonDescriptor)
	if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
		return errors.Wrap(err, "Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV steps have been run before]")
	}

	if addonDescriptor.TargetVectorID == "" {
		return errors.New("Parameter missing. Please provide the target vector id (e.g. by running step abapAddonAssemblyKitCreateTargetVector beforehand")
	}
	targetVector := new(aakaas.TargetVector)
	targetVector.InitExisting(addonDescriptor.TargetVectorID)

	switch config.TargetVectorScope {
	case string(aakaas.TargetVectorStatusTest):
		log.Entry().Infof("Publish target vector %s for test use", addonDescriptor.TargetVectorID)
	case string(aakaas.TargetVectorStatusProductive):
		log.Entry().Infof("Publish target vector %s for productive use", addonDescriptor.TargetVectorID)
	default:
		return errors.New("Invalid Value of configuration Parameter TargetVectorScope: " + config.TargetVectorScope)
	}

	if err := targetVector.PublishTargetVector(conn, aakaas.TargetVectorStatus(config.TargetVectorScope)); err != nil {
		return err
	}

	log.Entry().Info("Waiting for target vector publishing to finish")
	if err := targetVector.PollForStatus(conn, aakaas.TargetVectorStatus(config.TargetVectorScope)); err != nil {
		return err
	}

	log.Entry().Info("Success: Publishing finised")
	return nil
}
