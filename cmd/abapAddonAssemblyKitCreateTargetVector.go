package cmd

import (
	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitCreateTargetVector(config abapAddonAssemblyKitCreateTargetVectorOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitCreateTargetVector(&config, telemetryData, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCreateTargetVector(config *abapAddonAssemblyKitCreateTargetVectorOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender, cpe *abapAddonAssemblyKitCreateTargetVectorCommonPipelineEnvironment) error {
	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client); err != nil {
		return err
	}

	addonDescriptor := new(abaputils.AddonDescriptor)
	if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
		return errors.Wrap(err, "Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV steps have been run before]")
	}

	var tv aakaas.TargetVector
	if err := tv.InitNew(addonDescriptor); err != nil {
		return err
	}

	log.Entry().Infof("Create target vector for product %s version %s", addonDescriptor.AddonProduct, addonDescriptor.AddonVersionYAML)
	if err := tv.CreateTargetVector(conn); err != nil {
		return err
	}

	log.Entry().Infof("Created target vector %s", tv.ID)
	addonDescriptor.TargetVectorID = tv.ID

	log.Entry().Info("Write target vector to CommonPipelineEnvironment")
	cpe.abap.addonDescriptor = addonDescriptor.AsJSONstring()
	return nil
}
