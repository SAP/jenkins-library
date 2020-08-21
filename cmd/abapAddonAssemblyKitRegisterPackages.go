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

func abapAddonAssemblyKitRegisterPackages(config abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData) {
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
	err := runAbapAddonAssemblyKitRegisterPackages(&config, telemetryData, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitRegisterPackages(config *abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender) error {
	var repos []abaputils.Repository
	json.Unmarshal([]byte(config.Repositories), &repos)

	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	for _, repo := range repos {
		if repo.Status == "P" {
			filename := filepath.Base(repo.SarXMLFilePath)
			conn.Header["Content-Filename"] = []string{filename}
			sarFile, err := ioutil.ReadFile(repo.SarXMLFilePath)
			if err != nil {
				return err
			}
			err = conn.uploadSarFile("/odata/aas_file_upload", sarFile)
			if err != nil {
				return err
			}
		}
	}
	// we need a second connector without the added Header
	conn2 := new(connector)
	conn2.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	for _, repo := range repos {
		if repo.Status == "P" {
			var p pckg
			p.init(repo, *conn2)
			err := p.register()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *pckg) register() error {
	p.connector.getToken()
	appendum := "/odata/aas_ocs_package/RegisterPackage?Name='" + p.PackageName + "'"
	_, err := p.connector.post(appendum, "")
	if err != nil {
		return err
	}
	//TODO was kommt als return zurück? interessiert mich der return überhapt jenseits von fehler/kein fehler? vielleicht ändert sich der status? dann müsste es zurück ins cpe
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
