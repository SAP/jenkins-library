package cmd

import (
	"encoding/json"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
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

	log.Entry().Info(config.Username)
	log.Entry().Info(config.Password)
	var u string
	var p string
	u = config.Username
	p = config.Password
	log.Entry().Info(u)
	log.Entry().Info(p)

	for i := range addonDescriptor.Repositories {
		var p pckg
		p.init(addonDescriptor.Repositories[i], *conn)
		if addonDescriptor.Repositories[i].Status == "L" {
			err := p.release()
			if err != nil {
				return err
			}
			p.changeStatus(&addonDescriptor.Repositories[i])
		} else {
			log.Entry().Infof("Package %s has status %s, cannot release this package", p.PackageName, p.Status)
		}
	}
	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

// TODO loop wieder ausbauen
func (p *pckg) release() error {
	var body []byte
	var err error
	if p.PackageName == "" {
		return errors.New("Parameter missing. Please provide the name of the package which should be released")
	}
	log.Entry().Infof("Release package %s", p.PackageName)

	isReleased := false
	p.connector.getToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/ReleasePackage?Name='" + p.PackageName + "'"
	tryAgain := 0
	for !isReleased {
		body, err = p.connector.post(appendum, "")
		if err != nil {
			tryAgain = tryAgain + 1
			if tryAgain == 5 {
				return err
			}
			log.Entry().Info("Release did not work, let's try again in 15s")
			time.Sleep(15 * time.Second)
		} else {
			isReleased = true
		}
	}
	var jPck jsonPackageFromGet
	json.Unmarshal(body, &jPck)
	p.Status = jPck.Package.Status
	return nil
}
