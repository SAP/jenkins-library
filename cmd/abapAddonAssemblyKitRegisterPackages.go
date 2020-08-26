package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitRegisterPackages(config abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment) {
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
	err := runAbapAddonAssemblyKitRegisterPackages(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitRegisterPackages(config *abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment) error {
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	for i := range addonDescriptor.Repositories {
		if addonDescriptor.Repositories[i].Status == "P" {
			if addonDescriptor.Repositories[i].SarXMLFilePath == "" {
				return errors.New("Parameter missing. Please provide the path to the SAR file")
			}
			filename := filepath.Base(addonDescriptor.Repositories[i].SarXMLFilePath)
			conn.Header["Content-Filename"] = []string{filename}
			sarFile, err := ioutil.ReadFile(addonDescriptor.Repositories[i].SarXMLFilePath)
			if err != nil {
				return err
			}
			log.Entry().Infof("Upload SAR file %s", filename)
			err = conn.uploadSarFile("/odata/aas_file_upload", sarFile)
			if err != nil {
				return err
			}
		}
	}
	// we need a second connector without the added Header
	conn2 := new(connector)
	conn2.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	for i := range addonDescriptor.Repositories {
		var p pckg
		p.init(addonDescriptor.Repositories[i], *conn2)
		if addonDescriptor.Repositories[i].Status == "P" {
			err := p.register()
			if err != nil {
				return err
			}
			p.changeStatus(&addonDescriptor.Repositories[i])
		} else {
			log.Entry().Infof("Package %s has status %s, cannot register this package", p.PackageName, p.Status)
		}
	}
	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

func (p *pckg) changeStatus(initialRepo *abaputils.Repository) {
	initialRepo.Status = p.Status
}

func (p *pckg) register() error {
	if p.PackageName == "" {
		return errors.New("Parameter missing. Please provide the name of the package which should be registered")
	}
	log.Entry().Infof("Register package %s", p.PackageName)
	p.connector.getToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/RegisterPackage?Name='" + p.PackageName + "'"
	body, err := p.connector.post(appendum, "")
	if err != nil {
		return err
	}

	var jPck jsonPackageFromGet
	json.Unmarshal(body, &jPck)
	p.Status = jPck.Package.Status
	log.Entry().Infof("Package status %s", p.Status)
	return nil
}

func (conn connector) uploadSarFile(appendum string, sarFile []byte) error {
	url := conn.Baseurl + appendum
	response, err := conn.Client.SendRequest("PUT", url, bytes.NewBuffer(sarFile), conn.Header, nil)
	if err != nil {
		if response == nil {
			return errors.Wrap(err, "Upload of SAR file failed")
		}
		defer response.Body.Close()
		errorbody, _ := ioutil.ReadAll(response.Body)
		return errors.Wrapf(err, "Upload of SAR file failed: %v", string(errorbody))
	}
	defer response.Body.Close()
	return nil
}
