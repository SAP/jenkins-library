package cmd

import (
	"encoding/json"

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

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitPublishTargetVector(&config, telemetryData, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitPublishTargetVector(config *abapAddonAssemblyKitPublishTargetVectorOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender) error {
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	if addonDescriptor.TargetVectorID == "" {
		return errors.New("Parameter missing. Please provide the target vector id")
	}

	if config.ScopeTV == "T" {
		log.Entry().Infof("Publish target vector %s to test SPC", addonDescriptor.TargetVectorID)
	}
	if config.ScopeTV == "P" {
		log.Entry().Infof("Publish target vector %s to SPC", addonDescriptor.TargetVectorID)
	}
	conn.getToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/PublishTargetVector?Id='" + addonDescriptor.TargetVectorID + "'&Scope='" + config.ScopeTV + "'"
	_, err := conn.post(appendum, "")
	if err != nil {
		return err
	}
	return nil
}
