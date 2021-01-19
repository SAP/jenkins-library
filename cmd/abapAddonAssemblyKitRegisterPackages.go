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
			log.Entry().Infof("Trying to read file %s", repos[i].SarXMLFilePath)
			sarFile, err := readFileFunc(repos[i].SarXMLFilePath)
			if err != nil {
				return err
			}
			log.Entry().Infof("... %d bytes read", len(sarFile))
			if len(sarFile) == 0 {
				return errors.New("File has no content - 0 bytes")
			}
			log.Entry().Infof("Upload SAR file %s in chunks", filename)
			err = conn.UploadSarFileInChunks("/odata/aas_file_upload", filename, sarFile)
			if err != nil {
				return err
			}
			log.Entry().Info("...done")
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
			log.Entry().Infof("Trying to Register Package %s", pack.PackageName)
			err := pack.Register()
			if err != nil {
				return repos, err
			}
			log.Entry().Info("...done, take over new status")
			pack.ChangeStatus(&repos[i])
			log.Entry().Info("...done")
		} else {
			log.Entry().Infof("Package %s has status %s, cannot register this package", pack.PackageName, pack.Status)
		}
	}
	return repos, nil
}
