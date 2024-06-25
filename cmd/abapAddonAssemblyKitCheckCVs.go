package cmd

import (
	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitCheckCVs(config abapAddonAssemblyKitCheckCVsOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundle()
	telemetryData.BuildTool = "AAKaaS"
	if err := runAbapAddonAssemblyKitCheckCVs(&config, &utils, cpe); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheckCVs(config *abapAddonAssemblyKitCheckCVsOptions, utils *aakaas.AakUtils, cpe *abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment) error {

	log.Entry().Info("╔══════════════════════════════╗")
	log.Entry().Info("║ abapAddonAssemblyKitCheckCVs ║")
	log.Entry().Info("╚══════════════════════════════╝")

	conn := new(abapbuild.Connector)

	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, *utils, "", config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	log.Entry().Infof("Reading Product Version Information from addonDescriptor (aka addon.yml) file: %s", config.AddonDescriptorFileName)
	addonDescriptor, err := (*utils).ReadAddonDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return err
	}

	for i, repo := range addonDescriptor.Repositories {
		componentVersion := new(aakaas.ComponentVersion)
		if err := componentVersion.ConstructComponentVersion(addonDescriptor.Repositories[i], *conn); err != nil {
			return err
		}
		if err := componentVersion.Validate(); err != nil {
			return err
		}
		componentVersion.CopyVersionFieldsToRepo(&addonDescriptor.Repositories[i])

		log.Entry().Infof("Using cCTS %t", repo.UseClassicCTS)
		log.Entry().Infof("CommitId %s", repo.CommitID)

		if !repo.UseClassicCTS && repo.CommitID == "" {
			return errors.Errorf("CommitID missing in repo '%s' of the addon.yml", repo.Name)
		}
	}

	// now Software Component Versions fields are valid, but maybe Product Version was checked before, so copy that part from CPE
	// we don't care for errors
	// scenario 1: config.AddonDescriptor is empty since checkCVs is the first step in the pipeline, then the empty result is fine anyway
	// scenario 2: for some reason config.AddonDescriptor is corrupt - then we insert the valid data but delete the repositories which will ensure issue is found later on
	addonDescriptorCPE, _ := abaputils.ConstructAddonDescriptorFromJSON([]byte(config.AddonDescriptor))
	if len(addonDescriptorCPE.AddonProduct) == 0 {
		log.Entry().Info("No Product Version information present yet in the addonDescriptor of CommonPipelineEnvironment")
	} else {
		log.Entry().Infof("Information for Product Version %s taken from addonDescriptor of CommonPipelineEnvironment", addonDescriptorCPE.AddonProduct)
	}
	addonDescriptorCPE.SetRepositories(addonDescriptor.Repositories)
	cpe.abap.addonDescriptor = string(addonDescriptorCPE.AsJSON())
	log.Entry().Info("Wrote addonDescriptor to CommonPipelineEnvironment")
	return nil
}

// take the product part from CPE and the repositories part from the YAML file
func combineYAMLRepositoriesWithCPEProduct(addonDescriptor abaputils.AddonDescriptor, addonDescriptorFromCPE abaputils.AddonDescriptor) abaputils.AddonDescriptor {
	addonDescriptorFromCPE.Repositories = addonDescriptor.Repositories
	return addonDescriptorFromCPE
}
