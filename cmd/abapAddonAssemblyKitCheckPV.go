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

func abapAddonAssemblyKitCheckPV(config abapAddonAssemblyKitCheckPVOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitCheckPV(&config, telemetryData, &client, cpe, abaputils.ReadAddonDescriptor)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheckPV(config *abapAddonAssemblyKitCheckPVOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment, readAdoDescriptor abaputils.ReadAddonDescriptorType) error {
	var addonDescriptorFromCPE abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptorFromCPE)
	addonDescriptor, err := readAdoDescriptor(config.AddonDescriptorFileName)
	addonDescriptor = combineYAMLProductWithCPERepositories(addonDescriptor, addonDescriptorFromCPE)
	if err != nil {
		return err
	}
	conn := new(abapbuild.Connector)
	conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)

	var p productVersion
	p.init(addonDescriptor, *conn)
	err = p.validate()
	if err != nil {
		return err
	}
	p.copyFieldsToRepo(&addonDescriptor)
	log.Entry().Info("Write the resolved version to the CommonPipelineEnvironment")
	toCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(toCPE)
	return nil
}

func combineYAMLProductWithCPERepositories(addonDescriptor abaputils.AddonDescriptor, addonDescriptorFromCPE abaputils.AddonDescriptor) abaputils.AddonDescriptor {
	addonDescriptor.Repositories = addonDescriptorFromCPE.Repositories
	return addonDescriptor
}

func (p *productVersion) init(desc abaputils.AddonDescriptor, conn abapbuild.Connector) {
	p.Connector = conn
	p.Name = desc.AddonProduct
	p.VersionYAML = desc.AddonVersionYAML
}

func (p *productVersion) copyFieldsToRepo(initialAddonDescriptor *abaputils.AddonDescriptor) {
	initialAddonDescriptor.AddonVersion = p.Version
	initialAddonDescriptor.AddonSpsLevel = p.SpsLevel
	initialAddonDescriptor.AddonPatchLevel = p.PatchLevel
}

func (p *productVersion) validate() error {
	log.Entry().Infof("Validate product %s version %s and resolve version", p.Name, p.VersionYAML)
	appendum := "/odata/aas_ocs_package/ValidateProductVersion?Name='" + p.Name + "'&Version='" + p.VersionYAML + "'"
	body, err := p.Connector.Get(appendum)
	if err != nil {
		return err
	}
	var jPV jsonProductVersion
	json.Unmarshal(body, &jPV)
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
