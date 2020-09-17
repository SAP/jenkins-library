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

func abapAddonAssemblyKitCheckCVs(config abapAddonAssemblyKitCheckCVsOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitCheckCVs(&config, telemetryData, &client, cpe, abaputils.ReadAddonDescriptor)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheckCVs(config *abapAddonAssemblyKitCheckCVsOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment, readAdoDescriptor abaputils.ReadAddonDescriptorType) error {
	var addonDescriptorFromCPE abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptorFromCPE)
	addonDescriptor, err := readAdoDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return err
	}
	addonDescriptor = combineYAMLRepositoriesWithCPEProduct(addonDescriptor, addonDescriptorFromCPE)
	conn := new(abapbuild.Connector)
	conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)

	for i := range addonDescriptor.Repositories {
		var c componentVersion
		c.initCV(addonDescriptor.Repositories[i], *conn)
		err := c.validate()
		if err != nil {
			return err
		}
		c.copyFieldsToRepo(&addonDescriptor.Repositories[i])
	}
	log.Entry().Info("Write the resolved versions to the CommonPipelineEnvironment")
	toCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(toCPE)
	return nil
}

//take the product part from CPE and the repositories part from the YAML file
func combineYAMLRepositoriesWithCPEProduct(addonDescriptor abaputils.AddonDescriptor, addonDescriptorFromCPE abaputils.AddonDescriptor) abaputils.AddonDescriptor {
	addonDescriptorFromCPE.Repositories = addonDescriptor.Repositories
	return addonDescriptorFromCPE
}

func (c *componentVersion) initCV(repo abaputils.Repository, conn abapbuild.Connector) {
	c.Connector = conn
	c.Name = repo.Name
	c.VersionYAML = repo.VersionYAML
}

func (c *componentVersion) copyFieldsToRepo(initialRepo *abaputils.Repository) {
	initialRepo.Version = c.Version
	initialRepo.SpLevel = c.SpLevel
	initialRepo.PatchLevel = c.PatchLevel
}

func (c *componentVersion) validate() error {
	log.Entry().Infof("Validate component %s version %s and resolve version", c.Name, c.VersionYAML)
	appendum := "/odata/aas_ocs_package/ValidateComponentVersion?Name='" + c.Name + "'&Version='" + c.VersionYAML + "'"
	body, err := c.Connector.Get(appendum)
	if err != nil {
		return err
	}
	var jCV jsonComponentVersion
	json.Unmarshal(body, &jCV)
	c.Name = jCV.ComponentVersion.Name
	c.Version = jCV.ComponentVersion.Version
	c.SpLevel = jCV.ComponentVersion.SpLevel
	c.PatchLevel = jCV.ComponentVersion.PatchLevel
	log.Entry().Infof("Resolved version %s, splevel %s, patchlevel %s", c.Version, c.SpLevel, c.PatchLevel)
	return nil
}

type jsonComponentVersion struct {
	ComponentVersion *componentVersion `json:"d"`
}

type componentVersion struct {
	abapbuild.Connector
	Name        string `json:"Name"`
	VersionYAML string
	Version     string `json:"Version"`
	SpLevel     string `json:"SpLevel"`
	PatchLevel  string `json:"PatchLevel"`
}
