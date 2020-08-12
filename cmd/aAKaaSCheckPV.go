package cmd

import (
	"encoding/json"
	"net/http/cookiejar"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func aAKaaSCheckPV(config aAKaaSCheckPVOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *aAKaaSCheckPVCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAAKaaSCheckPV(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAAKaaSCheckPV(config *aAKaaSCheckPVOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, commonPipelineEnvironment *aAKaaSCheckPVCommonPipelineEnvironment) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	conn := new(connector)
	conn.initAAK(config.AAKaaSEndpoint, config.Username, config.Password, &piperhttp.Client{})

	p := pv{
		connector: *conn,
	}
	p.validate(*config)
	commonPipelineEnvironment.PVersion = p.Version
	commonPipelineEnvironment.PSpspLevel = p.SpsLevel
	commonPipelineEnvironment.PPatchLevel = p.PatchLevel
	return nil
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

type jsonPV struct {
	PV *pv `json:"d"`
}

type pv struct {
	connector
	Name       string `json:"Name"`
	Version    string `json:"Version"`
	SpsLevel   string `json:"SpsLevel"`
	PatchLevel string `json:"PatchLevel"`
}

func (p *pv) validate(options aAKaaSCheckPVOptions) error {
	appendum := "/ValidateProductVersion?Name='" + options.AddonProduct + "'&Version='" + options.AddonVersion + "'"
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
