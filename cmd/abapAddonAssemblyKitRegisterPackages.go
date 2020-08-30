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

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitRegisterPackages(&config, telemetryData, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitRegisterPackages(config *abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender, cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment) error {
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	err := uploadSarFiles(addonDescriptor.Repositories, *conn)
	if err != nil {
		return err
	}

	// we need a second connector without the added Header
	conn2 := new(connector)
	conn2.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	addonDescriptor.Repositories, err = registerPackages(addonDescriptor.Repositories, *conn2)
	if err != nil {
		return err
	}

	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

func uploadSarFiles(repos []abaputils.Repository, conn connector) error {
	for i := range repos {
		if repos[i].Status == "P" {
			if repos[i].SarXMLFilePath == "" {
				return errors.New("Parameter missing. Please provide the path to the SAR file")
			}
			filename := filepath.Base(repos[i].SarXMLFilePath)
			conn.Header["Content-Filename"] = []string{filename}
			sarFile, err := ioutil.ReadFile(repos[i].SarXMLFilePath)
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
	return nil
}

func registerPackages(repos []abaputils.Repository, conn connector) ([]abaputils.Repository, error) {
	for i := range repos {
		var p pckg
		p.init(repos[i], conn)
		if repos[i].Status == "P" {
			err := p.register()
			if err != nil {
				return repos, err
			}
			p.changeStatus(&repos[i])
		} else {
			log.Entry().Infof("Package %s has status %s, cannot register this package", p.PackageName, p.Status)
		}
	}
	return repos, nil
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

	var jPck jsonPackage
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
