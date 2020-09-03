package cmd

import (
	"encoding/json"
	"net/http/cookiejar"

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
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)

	var p pv
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

func (conn *connector) initAAK(aAKaaSEndpoint string, username string, password string, inputclient piperhttp.Sender) {
	conn.Client = inputclient
	conn.Header = make(map[string][]string)
	conn.Header["Accept"] = []string{"application/json"}
	conn.Header["Content-Type"] = []string{"application/json"}

	cookieJar, _ := cookiejar.New(nil)
	conn.Client.SetOptions(piperhttp.ClientOptions{
		Username:  username,
		Password:  password,
		CookieJar: cookieJar,
	})
	conn.Baseurl = aAKaaSEndpoint
}

func (p *pv) init(desc abaputils.AddonDescriptor, conn connector) {
	p.connector = conn
	p.Name = desc.AddonProduct
	p.VersionYAML = desc.AddonVersionYAML
}

func (p *pv) copyFieldsToRepo(initialAddonDescriptor *abaputils.AddonDescriptor) {
	initialAddonDescriptor.AddonVersion = p.Version
	initialAddonDescriptor.AddonSpsLevel = p.SpsLevel
	initialAddonDescriptor.AddonPatchLevel = p.PatchLevel
}

func (p *pv) validate() error {
	log.Entry().Infof("Validate product %s version %s and resolve version", p.Name, p.VersionYAML)
	appendum := "/odata/aas_ocs_package/ValidateProductVersion?Name='" + p.Name + "'&Version='" + p.VersionYAML + "'"
	body, err := p.connector.get(appendum)
	if err != nil {
		return err
	}
	var jPV jsonPV
	json.Unmarshal(body, &jPV)
	p.Name = jPV.PV.Name
	p.Version = jPV.PV.Version
	p.SpsLevel = jPV.PV.SpsLevel
	p.PatchLevel = jPV.PV.PatchLevel
	log.Entry().Infof("Resolved version %s, spslevel %s, patchlevel %s", p.Version, p.SpsLevel, p.PatchLevel)
	return nil
}

type jsonPV struct {
	PV *pv `json:"d"`
}

type pv struct {
	connector
	Name           string `json:"Name"`
	VersionYAML    string
	Version        string `json:"Version"`
	SpsLevel       string `json:"SpsLevel"`
	PatchLevel     string `json:"PatchLevel"`
	TargetVectorID string
}
