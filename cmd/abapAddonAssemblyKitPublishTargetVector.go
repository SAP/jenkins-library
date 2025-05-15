package cmd

import (
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitPublishTargetVector(config abapAddonAssemblyKitPublishTargetVectorOptions, telemetryData *telemetry.CustomData) {
	utils := aakaas.NewAakBundleWithTime(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervalInSeconds))
	telemetryData.BuildTool = "AAKaaS"

	if err := runAbapAddonAssemblyKitPublishTargetVector(&config, &utils); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitPublishTargetVector(config *abapAddonAssemblyKitPublishTargetVectorOptions, utils *aakaas.AakUtils) error {

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, *utils, config.AbapAddonAssemblyKitOriginHash, config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}
	conn.MaxRuntime = (*utils).GetMaxRuntime()
	conn.PollingInterval = (*utils).GetPollingInterval()

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

	log.Entry().Info("Success: Publishing finished")
	return nil
}
