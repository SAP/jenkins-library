package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitReleasePackages(config abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData) {
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
	err := runAbapAddonAssemblyKitReleasePackages(&config, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReleasePackages(config *abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) error {
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	var repos []abaputils.Repository
	json.Unmarshal([]byte(config.Repositories), &repos)

	for _, repo := range repos {
		var p pckg
		p.init(repo, *conn)
		err := p.release()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *pckg) release() error {
	p.connector.getToken()
	appendum := "/ReleasePackage?Name='" + p.PackageName + "'"
	_, err := p.connector.post(appendum, "")
	if err != nil {
		return err
	}
	//TODO was kommt als return zurück? interessiert mich der return überhapt jenseits von fehler/kein fehler?
	return nil
}
