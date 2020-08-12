package cmd

import (
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func aAKaaSRegisterDeliveryTransport(config aAKaaSRegisterDeliveryTransportOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAAKaaSRegisterDeliveryTransport(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAAKaaSRegisterDeliveryTransport(config *aAKaaSRegisterDeliveryTransportOptions, telemetryData *telemetry.CustomData, command command.ExecRunner) error {
	conn := new(connector)
	conn.initAAK(config.AAKaaSEndpoint, config.Username, config.Password, &piperhttp.Client{})
	p := pckg{
		connector: *conn,
		Name:      config.PackageName,
	}
	//TODO https://wiki.wdf.sap.corp/wiki/pages/viewpage.action?spaceKey=A4H&title=Build+Pipeline+for+Partner+Addons da steht noch was von upload file, ist dass das sar file?
	// Wie sieht der aufruf genau aus?
	// dann müsste ich als input für den schritt noch das sarfile dazu fügen
	p.register()
	return nil
}

func (p *pckg) register() error {
	p.connector.getToken()
	appendum := "/RegisterPackage?Name='" + p.Name + "'"
	// body, err := p.connector.post2(appendum, "")
	_, err := p.connector.post2(appendum, "")
	if err != nil {
		return err
	}
	//TODO was kommt als return zurück?
	// var jPck jsonPackage
	// json.Unmarshal(body, &jPck)
	// p.Name = jPck.Package.Name
	// p.Type = jPck.Package.Type
	// p.PreviousDeliveryCommit = jPck.Package.PreviousDeliveryCommit
	// p.Status = jPck.Package.Status
	return nil
}
