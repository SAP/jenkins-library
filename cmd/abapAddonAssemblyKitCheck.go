package cmd

import (
	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitCheck(config abapAddonAssemblyKitCheckOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *abapAddonAssemblyKitCheckCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundle()

	err := runAbapAddonAssemblyKitCheck(&config, telemetryData, utils, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheck(config *abapAddonAssemblyKitCheckOptions, telemetryData *telemetry.CustomData, utils aakaas.AakUtils, commonPipelineEnvironment *abapAddonAssemblyKitCheckCommonPipelineEnvironment) error {

	log.Entry().Info("╔═══════════════════════════╗")
	log.Entry().Info("║ abapAddonAssemblyKitCheck ║")
	log.Entry().Info("╚═══════════════════════════╝")

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, utils, "", config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	log.Entry().Infof("reading addonDescriptor (aka addon.yml) file: %s", config.AddonDescriptorFileName)
	addonDescriptor, err := utils.ReadAddonDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return err
	}
	log.Entry().Info("building product modelling (and resolving potential wildcards)")
	pvh, err := aakaas.NewProductVersionHeader(&addonDescriptor, conn)
	if err != nil {
		return err
	}
	printProductVersionHeader(*pvh)

	log.Entry().Info("calling AAKaaS to check product modelling...")
	if err := pvh.CheckAndResolveVersion(conn); err != nil {
		return err
	}
	log.Entry().Info("... success!")
	pvh.SyncAddonDescriptorVersionFields(&addonDescriptor)
	log.Entry().Info("resolved version fields:")
	printAddonDescriptorVersionFields(addonDescriptor)
	log.Entry().Info("transferring addonDescriptor to commonPipelineEnvironment for usage by subsequent steps of the pipeline")
	commonPipelineEnvironment.abap.addonDescriptor = string(addonDescriptor.AsJSON())

	publishAddonYaml(config, utils)
	return nil
}

func printProductVersionHeader(pvh aakaas.ProductVersionHeader) {
	logLine30 := "──────────────────────────────"
	log.Entry().Infof("┌─%-30v─┬─%-30v─┐", logLine30, logLine30)
	log.Entry().Infof("│ %-30v │ %-30v │", "Product Name", pvh.ProductName)
	log.Entry().Infof("│ %-30v │ %-30v │", "Product Version", pvh.SemanticProductVersion)
	log.Entry().Infof("├─%-30v─┼─%-30v─┤", logLine30, logLine30)
	log.Entry().Infof("│ %-30v │ %-30v │", "Software Component Name", "Software Component Version")
	log.Entry().Infof("├─%-30v─┼─%-30v─┤", logLine30, logLine30)
	for _, pvc := range pvh.Content {
		log.Entry().Infof("│ %-30v │ %-30v │", pvc.SoftwareComponentName, pvc.SemanticSoftwareComponentVersion)
	}
	log.Entry().Infof("└─%-30v─┴─%-30v─┘", logLine30, logLine30)
}

func printAddonDescriptorVersionFields(addonDescriptor abaputils.AddonDescriptor) {
	logLine30 := "──────────────────────────────"
	logLine4 := "────"
	log.Entry().Infof("┌─%-30v─┬─%-4v─┬─%-4v─┬─%-4v─┐", logLine30, logLine4, logLine4, logLine4)
	log.Entry().Infof("│ %-30v │ %-4v │ %-4v │ %-4v │", "Name", "Vers", "SP", "Pat.")
	log.Entry().Infof("├─%-30v─┼─%-4v─┼─%-4v─┼─%-4v─┤", logLine30, logLine4, logLine4, logLine4)
	log.Entry().Infof("│ %-30v │ %-4v │ %-4v │ %-4v │", addonDescriptor.AddonProduct, addonDescriptor.AddonVersion, addonDescriptor.AddonSpsLevel, addonDescriptor.AddonPatchLevel)
	for _, repo := range addonDescriptor.Repositories {
		log.Entry().Infof("│ %-30v │ %-4v │ %-4v │ %-4v │", repo.Name, repo.Version, repo.SpLevel, repo.PatchLevel)
	}
	log.Entry().Infof("└─%-30v─┴─%-4v─┴─%-4v─┴─%-4v─┘", logLine30, logLine4, logLine4, logLine4)
}

func publishAddonYaml(config *abapAddonAssemblyKitCheckOptions, utils aakaas.AakUtils) {
	var filesToPublish []piperutils.Path
	log.Entry().Infof("adding %s to be published", config.AddonDescriptorFileName)
	filesToPublish = append(filesToPublish, piperutils.Path{Target: config.AddonDescriptorFileName, Name: "AddonDescriptor", Mandatory: true})
	log.Entry().Infof("publishing %v files", len(filesToPublish))
	if err := piperutils.PersistReportsAndLinks("abapAddonAssemblyKitCheckPV", "", utils, filesToPublish, nil); err != nil {
		log.Entry().WithError(err).Error("failed to persist report information")
	}
}
