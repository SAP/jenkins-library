package cmd

import (
	"encoding/json"
	"net/url"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitCheckPV(config abapAddonAssemblyKitCheckPVOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundle()
	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	if err := runAbapAddonAssemblyKitCheckPV(&config, telemetryData, utils, cpe); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}
func runAbapAddonAssemblyKitCheckPV(config *abapAddonAssemblyKitCheckPVOptions, telemetryData *telemetry.CustomData, utils aakaas.AakUtils, cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) error {

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, utils); err != nil {
		return err
	}

	log.Entry().Infof("Reading Product Version Information from addonDescriptor (aka addon.yml) file: %s", config.AddonDescriptorFileName)
	addonDescriptor, err := utils.ReadAddonDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return err
	}

	pv := new(productVersion).init(addonDescriptor, *conn)
	err = pv.validateAndResolveVersionFields()
	if err != nil {
		return err
	}
	pv.transferVersionFields(&addonDescriptor)

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

func (p *productVersion) init(desc abaputils.AddonDescriptor, conn abapbuild.Connector) *productVersion {
	p.Connector = conn
	p.Name = desc.AddonProduct
	p.VersionYAML = desc.AddonVersionYAML

	return p
}

func (p *productVersion) transferVersionFields(initialAddonDescriptor *abaputils.AddonDescriptor) {
	initialAddonDescriptor.AddonVersion = p.Version
	initialAddonDescriptor.AddonSpsLevel = p.SpsLevel
	initialAddonDescriptor.AddonPatchLevel = p.PatchLevel
}

func (p *productVersion) validateAndResolveVersionFields() error {
	log.Entry().Infof("Validate product '%s' version '%s' and resolve version", p.Name, p.VersionYAML)
	appendum := "/odata/aas_ocs_package/ValidateProductVersion?Name='" + url.QueryEscape(p.Name) + "'&Version='" + url.QueryEscape(p.VersionYAML) + "'"
	body, err := p.Connector.Get(appendum)
	if err != nil {
		return err
	}
	var jPV jsonProductVersion
	if err := json.Unmarshal(body, &jPV); err != nil {
		return errors.Wrap(err, "Unexpected AAKaaS response for Validate Product Version: "+string(body))
	}
	p.Name = jPV.ProductVersion.Name
	p.Version = jPV.ProductVersion.Version
	p.SpsLevel = jPV.ProductVersion.SpsLevel
	p.PatchLevel = jPV.ProductVersion.PatchLevel
	log.Entry().Infof("Resolved version %s, spslevel %s, patchlevel %s", p.Version, p.SpsLevel, p.PatchLevel)
	return nil
}

type jsonProductVersion struct {
	ProductVersion *productVersion `json:"d"`
}

type productVersion struct {
	abapbuild.Connector
	Name           string `json:"Name"`
	VersionYAML    string
	Version        string `json:"Version"`
	SpsLevel       string `json:"SpsLevel"`
	PatchLevel     string `json:"PatchLevel"`
	TargetVectorID string
}
