package cmd

import (
	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitCheckPV(config abapAddonAssemblyKitCheckPVOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundle()
	telemetryData.BuildTool = "AAKaaS"
	if err := runAbapAddonAssemblyKitCheckPV(&config, utils, cpe); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}
func runAbapAddonAssemblyKitCheckPV(config *abapAddonAssemblyKitCheckPVOptions, utils aakaas.AakUtils, cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) error {

	log.Entry().Info("╔═════════════════════════════╗")
	log.Entry().Info("║ abapAddonAssemblyKitCheckPV ║")
	log.Entry().Info("╚═════════════════════════════╝")

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, utils, "", config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	log.Entry().Infof("Reading Product Version Information from addonDescriptor (aka addon.yml) file: %s", config.AddonDescriptorFileName)
	addonDescriptor, err := utils.ReadAddonDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return err
	}

	pv := new(aakaas.ProductVersion)
	if err := pv.ConstructProductversion(addonDescriptor, *conn); err != nil {
		return err
	}
	if err = pv.ValidateAndResolveVersionFields(); err != nil {
		return err
	}
	pv.CopyVersionFieldsToDescriptor(&addonDescriptor)

	// now Product Version fields are valid, but maybe Component Versions (Repositories) were checked before, so copy that part from CPE
	// we don't care for errors
	// scenario 1: config.AddonDescriptor is empty since checkPV is the first step in the pipeline, then the empty result is fine anyway
	// scenario 2: for some reason config.AddonDescriptor is corrupt - then we insert the valid data but delete the repositories which will ensure issue is found later on
	addonDescriptorCPE, _ := abaputils.ConstructAddonDescriptorFromJSON([]byte(config.AddonDescriptor))
	if len(addonDescriptorCPE.Repositories) == 0 {
		log.Entry().Info("No Software Component Information present yet in the addonDescriptor of CommonPipelineEnvironment")
	} else {
		log.Entry().Infof("Information for %v Software Component Repositories taken from addonDescriptor of CommonPipelineEnvironment", len(addonDescriptorCPE.Repositories))
	}
	addonDescriptor.SetRepositories(addonDescriptorCPE.Repositories)
	cpe.abap.addonDescriptor = string(addonDescriptor.AsJSON())
	log.Entry().Info("Wrote addonDescriptor to CommonPipelineEnvironment")

	var filesToPublish []piperutils.Path
	log.Entry().Infof("Add %s to be published", config.AddonDescriptorFileName)
	filesToPublish = append(filesToPublish, piperutils.Path{Target: config.AddonDescriptorFileName, Name: "AddonDescriptor", Mandatory: true})
	log.Entry().Infof("Publishing %v files", len(filesToPublish))
	if err := piperutils.PersistReportsAndLinks("abapAddonAssemblyKitCheckPV", "", utils, filesToPublish, nil); err != nil {
		log.Entry().WithError(err).Error("failed to persist report information")
	}

	return nil
}
