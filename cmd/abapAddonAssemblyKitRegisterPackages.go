package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	var repos []abaputils.Repository
	json.Unmarshal([]byte(config.Repositories), &repos)

	//TODO https://wiki.wdf.sap.corp/wiki/pages/viewpage.action?spaceKey=A4H&title=Build+Pipeline+for+Partner+Addons da steht noch was von upload file, ist dass das sar file?
	// Wie sieht der aufruf genau aus?
	// dann müsste ich als input für den schritt noch das sarfile dazu fügen
	// for _, repo := range repos {
	// 	var p pckg
	// 	p.init(repo, *conn)
	// 	err := p.register()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for _, repo := range repos {
		//
		filename := filepath.Base(repo.SarXMLFilePath)
		fmt.Println("filename " + filename)
		// var contDisp string
		// TODO nimmt er das mit den ' statt " ?
		// contDisp = "form-data; name='file'; filename='" + filename + "'"
		// fmt.Println("content-disposition " + contDisp)
		// conn.Header["Content-Disposition"] = []string{contDisp}

		conn.Header["Content-Filename"] = []string{filename}
		sarFile, err := ioutil.ReadFile(repo.SarXMLFilePath)
		if err != nil {
			return err
		}
		url := "https://w7q.dmzwdf.sap.corp/odata/aas_file_upload"
		_, err = conn.uploadSarFile(url, sarFile)
		if err != nil {
			return err
		}
	}

	// TODO register nach upload! am besten den connector nochmal neu aufsetzen
	return nil
}

func (p *pckg) register() error {
	p.connector.getToken()
	appendum := "/odata/aas_ocs_package/RegisterPackage?Name='" + p.PackageName + "'"
	_, err := p.connector.post(appendum, "")
	if err != nil {
		return err
	}
	//TODO was kommt als return zurück? interessiert mich der return überhapt jenseits von fehler/kein fehler?
	return nil
}

// TODO error messages
func (conn connector) uploadSarFile(url string, sarFile []byte) ([]byte, error) {
	response, err := conn.Client.SendRequest("PUT", url, bytes.NewBuffer(sarFile), conn.Header, nil)
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
