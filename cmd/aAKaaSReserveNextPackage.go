package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func aAKaaSReserveNextPackage(config aAKaaSReserveNextPackageOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *aAKaaSReserveNextPackageCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAAKaaSReserveNextPackage(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAAKaaSReserveNextPackage(config *aAKaaSReserveNextPackageOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, commonPipelineEnvironment *aAKaaSReserveNextPackageCommonPipelineEnvironment) error {
	conn := new(connector)
	conn.initAAK(config.AAKaaSEndpoint, config.Username, config.Password, &piperhttp.Client{})

	p := pckg{
		connector: *conn,
	}
	//TODO soll hier gepollt werden auf den status? ist bei dem ersten reserve next package und dann get das get leer?
	//TODO laut Dirk soll wenn das Paket nicht status P hat der Assembly schritt nicht ausgeführt werden => soll ich diese info per commonEnv weitergeben und der assembly step dann
	// nicht ausgeführt werden? oder soll das irgendwie in dieser großen pipeline auseinander gesteuert werden
	p.reserveNext(*config)
	p.get()
	commonPipelineEnvironment.PackageName = p.Name
	commonPipelineEnvironment.PackageType = p.Type
	commonPipelineEnvironment.PreviousDeliveryCommit = p.PreviousDeliveryCommit
	commonPipelineEnvironment.Namespace = p.Namespace
	return nil
}

type jsonPackage struct {
	Package *pckg `json:"d"`
}

type pckg struct {
	connector
	Name                   string `json:"Name"`
	Type                   string `json:"Type"`
	PreviousDeliveryCommit string `json:"PredecessorCommitId"`
	Status                 string `json:"Status"`
	Namespace              string `json:"Namespace"`
}

func (p *pckg) reserveNext(options aAKaaSReserveNextPackageOptions) error {
	p.connector.getToken()
	appendum := "/DeterminePackageForScv?Name='" + options.AddonComponent + "'&Version='" + options.AddonComponentVersion + "'"
	body, err := p.connector.post2(appendum, "")
	if err != nil {
		return err
	}
	var jPck jsonPackage
	json.Unmarshal(body, &jPck)
	p.Name = jPck.Package.Name
	p.Type = jPck.Package.Type
	p.PreviousDeliveryCommit = jPck.Package.PreviousDeliveryCommit
	p.Status = jPck.Package.Status
	return nil
}

func (p *pckg) get() error {
	appendum := "OcsPackageSet('" + p.Name + "')"
	body, err := p.connector.get(appendum)
	if err != nil {
		return err
	}
	var jPck jsonPackage
	json.Unmarshal(body, &jPck)
	p.Status = jPck.Package.Status
	p.Namespace = jPck.Package.Namespace
	return nil
}

//TODO das sollte irgendwie zum reuse gepackt werden und auch mit dem post von dem assembly step zusammengeführt werden
func (conn connector) post2(appendum string, importBody string) ([]byte, error) {
	url := conn.Baseurl + appendum
	var response *http.Response
	var err error
	if importBody != "" {
		response, err = conn.Client.SendRequest("POST", url, nil, conn.Header, nil)
	} else {
		response, err = conn.Client.SendRequest("POST", url, bytes.NewBuffer([]byte(importBody)), conn.Header, nil)
	}
	if err != nil {
		if response == nil {
			return nil, errors.Wrap(err, "Post failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errorbody, errors.Wrapf(err, "Post failed: %v", string(errorbody))
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return body, err
}
