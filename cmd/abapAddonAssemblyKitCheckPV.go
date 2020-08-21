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

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}
	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitCheckPV(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitCheckPV(config *abapAddonAssemblyKitCheckPVOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapAddonAssemblyKitCheckPVCommonPipelineEnvironment) error {
	var addonDescriptorFromCPE abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptorFromCPE)
	addonDescriptor, err := abaputils.ReadAddonDescriptor(config.AddonDescriptorFileName)
	addonDescriptor.Repositories = addonDescriptorFromCPE.Repositories
	if err != nil {
		return nil
	}

	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})

	var p pv
	p.init(addonDescriptor, *conn)
	err = p.validate()
	if err != nil {
		return err
	}
	addonDescriptor = p.convert()
	toCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(toCPE)
	return nil
}

// *******************************************************************************************************************************
// ************************************************************ REUSE ************************************************************
// *******************************************************************************************************************************

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

// TODO nochmal checken wie das mit dem verschiedenen Versionen ist => hab leider kein feld in das abaputils dafür eingebaut, geht es auch so?
func (p *pv) init(desc abaputils.AddonDescriptor, conn connector) {
	p.connector = conn
	p.Name = desc.AddonProduct
	p.VersionYAML = desc.AddonVersion
}

// TODO auf die verschiedenen versions achten!
func (p *pv) convert() abaputils.AddonDescriptor {
	var desc abaputils.AddonDescriptor
	desc.AddonProduct = p.Name
	desc.AddonVersion = p.Version
	desc.AddonSpsLevel = p.SpsLevel
	desc.AddonPatchLevel = p.PatchLevel
	desc.TargetVectorID = p.TargetVectorID
	return desc
}
func (p *pv) validate() error {
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
	return nil
}

type jsonPV struct {
	PV *pv `json:"d"`
}

// TODO TargetVectorID json string
type pv struct {
	connector
	Name           string `json:"Name"`
	VersionYAML    string
	Version        string `json:"Version"`
	SpsLevel       string `json:"SpsLevel"`
	PatchLevel     string `json:"PatchLevel"`
	TargetVectorID string
}
