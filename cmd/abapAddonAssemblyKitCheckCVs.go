package cmd

import (
	"encoding/json"

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
	err := runAbapAddonAssemblyKitCheckCVs(&config, telemetryData, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheckCVs(config *abapAddonAssemblyKitCheckCVsOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender, cpe *abapAddonAssemblyKitCheckCVsCommonPipelineEnvironment) error {
	var addonDescriptorFromCPE abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptorFromCPE)
	addonDescriptor, err := abaputils.ReadAddonDescriptor(config.AddonDescriptorFileName)
	if err != nil {
		return nil
	}
	addonDescriptor = combineYAMLRepositoriesWithCPEProduct(addonDescriptor, addonDescriptorFromCPE)
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})

	for i := range addonDescriptor.Repositories {
		var c cv
		c.init(addonDescriptor.Repositories[i], *conn)
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

func (c *cv) init(repo abaputils.Repository, conn connector) {
	c.connector = conn
	c.Name = repo.Name
	c.VersionYAML = repo.VersionYAML
}

func (c *cv) copyFieldsToRepo(initialRepo *abaputils.Repository) {
	initialRepo.Version = c.Version
	initialRepo.SpLevel = c.SpLevel
	initialRepo.PatchLevel = c.PatchLevel
}

func (c *cv) validate() error {
	log.Entry().Infof("Validate component %s version %s and resolve version", c.Name, c.VersionYAML)
	appendum := "/odata/aas_ocs_package/ValidateComponentVersion?Name='" + c.Name + "'&Version='" + c.VersionYAML + "'"
	body, err := c.connector.get(appendum)
	if err != nil {
		return err
	}
	var jCV jsonCV
	json.Unmarshal(body, &jCV)
	c.Name = jCV.CV.Name
	c.Version = jCV.CV.Version
	c.SpLevel = jCV.CV.SpLevel
	c.PatchLevel = jCV.CV.PatchLevel
	log.Entry().Infof("Resolved version %s, splevel %s, patchlevel %s", c.Version, c.SpLevel, c.PatchLevel)
	return nil
}

type jsonCV struct {
	CV *cv `json:"d"`
}

type cv struct {
	connector
	Name        string `json:"Name"`
	VersionYAML string
	Version     string `json:"Version"`
	SpLevel     string `json:"SpLevel"`
	PatchLevel  string `json:"PatchLevel"`
}
