package cmd

import (
	"encoding/json"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitRegisterPackages(config abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitRegisterPackages(&config, telemetryData, &client, cpe, reader)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitRegisterPackages(config *abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment, fileReader readFile) error {
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	conn := new(abapbuild.Connector)
	conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)
	err := uploadSarFiles(addonDescriptor.Repositories, *conn, fileReader)
	if err != nil {
		return err
	}

	// we need a second connector without the added Header
	conn2 := new(abapbuild.Connector)
	conn2.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)
	addonDescriptor.Repositories, err = registerPackages(addonDescriptor.Repositories, *conn2)
	if err != nil {
		return err
	}

	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}
