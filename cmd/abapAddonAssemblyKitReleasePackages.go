package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitReleasePackages(config abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment) {
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
	err := runAbapAddonAssemblyKitReleasePackages(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReleasePackages(config *abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment) error {
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	for i := range addonDescriptor.Repositories {
		if addonDescriptor.Repositories[i].Status == "L" {
			var p pckg
			p.init(addonDescriptor.Repositories[i], *conn)
			err := p.release()
			if err != nil {
				return err
			}
			p.changeStatus(&addonDescriptor.Repositories[i])
		}
	}
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

func (p *pckg) release() error {
	p.connector.getToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/ReleasePackage?Name='" + p.PackageName + "'"
	fmt.Println("send to " + appendum)
	body, err := p.connector.post(appendum, "")
	if err != nil {
		fmt.Println("error occured")
		return err
	}
	var jPck jsonPackageFromGet
	json.Unmarshal(body, &jPck)
	p.Status = jPck.Package.Status
	fmt.Println(p.Status)
	//TODO was kommt als return zurück? interessiert mich der return überhapt jenseits von fehler/kein fehler? vielleicht ändert sich der status? dann müsste es zurück ins cpe
	return nil
}
