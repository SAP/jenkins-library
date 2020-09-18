package cmd

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
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
	err := runAbapAddonAssemblyKitRegisterPackages(&config, telemetryData, &client, cpe, reader)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitRegisterPackages(config *abapAddonAssemblyKitRegisterPackagesOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitRegisterPackagesCommonPipelineEnvironment, fileReader readFile) error {
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	conn := new(abapbuild.Connector)
	conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)
	err := uploadSarFiles(addonDescriptor.Repositories, *conn, fileReader)
	if err != nil {
		return err
	}

	// we need a second connector without the added Header
	conn2 := new(abapbuild.Connector)
	conn2.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)
	addonDescriptor.Repositories, err = registerPackages(addonDescriptor.Repositories, *conn2)
	if err != nil {
		return err
	}

	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

func uploadSarFiles(repos []abaputils.Repository, conn abapbuild.Connector, readFileFunc readFile) error {
	for i := range repos {
		if repos[i].Status == string(aakaas.PackageStatusPlanned) {
			if repos[i].SarXMLFilePath == "" {
				return errors.New("Parameter missing. Please provide the path to the SAR file")
			}
			filename := filepath.Base(repos[i].SarXMLFilePath)
			conn.Header["Content-Filename"] = []string{filename}
			sarFile, err := readFileFunc(repos[i].SarXMLFilePath)
			if err != nil {
				return err
			}
			log.Entry().Infof("Upload SAR file %s", filename)
			err = conn.UploadSarFile("/odata/aas_file_upload", sarFile)
			if err != nil {
				return err
			}
		} else {
			log.Entry().Infof("Package %s has status %s, cannot upload the SAR file of this package", repos[i].PackageName, repos[i].Status)
		}
	}
	return nil
}

// for moocking
type readFile func(path string) ([]byte, error)

func reader(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func registerPackages(repos []abaputils.Repository, conn abapbuild.Connector) ([]abaputils.Repository, error) {
	for i := range repos {
		var pack aakaas.Package
		pack.InitPackage(repos[i], conn)
		if repos[i].Status == string(aakaas.PackageStatusPlanned) {
			err := pack.Register()
			if err != nil {
				return repos, err
			}
			pack.ChangeStatus(&repos[i])
		} else {
			log.Entry().Infof("Package %s has status %s, cannot register this package", pack.PackageName, pack.Status)
		}
	}
	return repos, nil
}
